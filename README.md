# Metadata Center

A near real-time load metric collection component, designed for intelligent inference scheduler in large-scale inference services.

English | [中文](README_ZH.md)

## Status

Early & quick developing

## Background

Load metrics is very import for LLM inference scheduler.

Typically, the following four load metrics are very important: (for each engine level)

1. Total number of requests
2. Token usage (KVCache usage)
3. Number of requests in Prefill
4. Prompt length in Prefill

Timeliness is critical in large scale service.
Poor timeliness will lead to large races, may choosing the same inference engine before the load metrics are updated.

There will be a fixed periodic delay, when polling metrics from engines.
Especially in large-scale scenarios, as the QPS (throughput) increases, the race will also increase significantly.

## Architecture

[![Architecture](docs/images/architecture.png)](docs/images/architecture.png)

Cooperating with Inference Gateway(i.e. [AIGW](https://github.com/aigw-project/aigw)), we can achieve near real-time load metric collection by the following steps:

1. Request proxy to Inference Engine:

    a. prefill & total request number: `+1`

    b. prefill prompt length: `+prompt-length`

2. First token responded

   a. prefill request number: `-1`

   b. prefill prompt length: `-prompt-length`

3. Request done

   a. total request number: `-1`

Even more, we can introduce CAS API to reduce race, when it is required in the feature.

## 📚 Documentation

- [Developer Guide](docs/en/developer_guide.md)
- [API Documentation](docs/en/api.md)
- [Roadmap](docs/en/ROADMAP.md)

## 📜 License

This project is licensed under [Apache 2.0](LICENSE).