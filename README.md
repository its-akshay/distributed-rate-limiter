# Distributed Rate Limiter

A distributed rate limiting service built using Go, Redis, PostgreSQL, and Kubernetes.

## Features
- Rule management APIs
- Redis-backed rate limiting
- PostgreSQL persistence
- REST APIs
- Docker support
- Kubernetes deployment (WIP)

## Tech Stack
- Go
- Gin
- PostgreSQL
- Redis
- Docker
- Kubernetes


## Grafana Dashboard

The project includes a pre-built Grafana dashboard.
Import:
grafana/dashboards/rate-limiter-dashboard.json
after configuring Prometheus as the data source.