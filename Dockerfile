# Copyright The AIGW Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Single stage build for simplicity
ARG DOCKER_MIRROR
FROM ${DOCKER_MIRROR}golang:1.23-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o metadata-center ./cmd/main.go

# Create logs directory
RUN mkdir -p logs

# Expose ports
EXPOSE 8080 8081

# Run the application
CMD ["./metadata-center", "run", "--config", "configs/config.toml"]