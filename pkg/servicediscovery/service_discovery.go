// Copyright The AIGW Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package servicediscovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/servicediscovery/types"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// DNSConfig contains configuration parameters for DNS-based service discovery.
type DNSConfig struct {
	Domain         string              // Domain name to perform DNS lookups on
	LookupInterval time.Duration       // Interval between DNS lookups
	GetLocalHosts  types.GetLocalHostsFunc // Function to get local host addresses for exclusion
}

// dnsServiceDiscovery implements DNS-based service discovery.
// It periodically performs DNS lookups to discover service instances.
type dnsServiceDiscovery struct {
	config    DNSConfig          // Configuration parameters
	mutex     sync.RWMutex       // Protects concurrent access to nodeList and hosts
	nodeList  map[string]struct{} // Set of discovered host IP addresses
	hosts     []string           // List of discovered hosts (excluding local host)
	localHost string             // Local host address to exclude from results
}

// NewDNSDiscovery creates a new DNS-based service discovery instance.
// It initializes the discovery service and starts the background lookup loop.
// Returns an error if local host detection fails.
func NewDNSDiscovery(config DNSConfig) (types.ServiceDiscovery, error) {
	var localHost string
	var err error
	if config.GetLocalHosts != nil {
		localHost, err = config.GetLocalHosts()
		if err != nil {
			return nil, fmt.Errorf("failed to get local hosts during discovery initialization: %w", err)
		}

		logger.Infof("Local hosts for exclusion: %s", localHost)
	}

	sd := &dnsServiceDiscovery{
		config:    config,
		nodeList:  make(map[string]struct{}),
		localHost: localHost,
	}

	sd.start()

	return sd, nil
}

// start begins the background DNS lookup loop.
// It runs in a separate goroutine and periodically performs DNS lookups.
func (sd *dnsServiceDiscovery) start() {
	ticker := time.NewTicker(sd.config.LookupInterval)
	logger.Infof("DNSLookUp loop started, interval: %s", sd.config.LookupInterval)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Recovered from panic in DNS lookup goroutine: %v", r)
			}
			ticker.Stop()
			logger.Infof("DNS lookup loop for domain '%s' has stopped.", sd.config.Domain)
		}()

		for range ticker.C {
			sd.dnsLookUp()
		}
	}()
}

// dnsLookUp performs a DNS lookup and updates the host list
// Removes stale hosts and adds new ones, updates metrics
func (sd *dnsServiceDiscovery) dnsLookUp() {
	hosts, err := net.LookupIP(sd.config.Domain)
	if err != nil {
		logger.Errorf("LookupIP failed for domain %s, err: %v", sd.config.Domain, err)
		return
	}

	logger.Infof("DNS lookup for domain %s found hosts: %v", sd.config.Domain, hosts)
	newHosts := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		newHosts[host.String()] = struct{}{}
	}

	if len(hosts) == 0 {
		logger.Errorf("DNS lookup returned empty result for domain %s", sd.config.Domain)
	}

	sd.mutex.Lock()
	defer sd.mutex.Unlock()

	for oldHost := range sd.nodeList {
		if _, exists := newHosts[oldHost]; !exists {
			delete(sd.nodeList, oldHost)
			prom.DNSLookupHosts.DeleteLabelValues(sd.config.Domain, oldHost)
			logger.Infof("Host removed for domain %s: %s, removed metrics", sd.config.Domain, oldHost)
		}
	}

	for newHost := range newHosts {
		if _, exists := sd.nodeList[newHost]; !exists {
			sd.nodeList[newHost] = struct{}{}
			prom.DNSLookupHosts.WithLabelValues(sd.config.Domain, newHost).Set(1)
			logger.Infof("Found new host for domain %s: %s, added metrics", sd.config.Domain, newHost)
		}
	}

	updatedHosts := make([]string, 0, len(sd.nodeList))
	for host := range sd.nodeList {
		if host == sd.localHost {
			continue
		}
		updatedHosts = append(updatedHosts, host)
	}
	sd.hosts = updatedHosts
}

// GetHosts returns the list of discovered service hosts.
// The list excludes the local host address and is thread-safe.
func (sd *dnsServiceDiscovery) GetHosts() []string {
	sd.mutex.RLock()
	defer sd.mutex.RUnlock()

	return sd.hosts
}
