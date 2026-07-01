# Distributed Rate Limiter

A production-ready distributed rate limiter built with **Go**, **Redis**, **PostgreSQL**, and **Kubernetes**. The project implements multiple rate limiting algorithms, exposes REST APIs, supports atomic operations using Redis Lua scripts, and provides observability through Prometheus and Grafana.

---

## Features

* Multiple rate limiting algorithms

  * Fixed Window
  * Sliding Window
  * Atomic Sliding Window (Redis Lua Script)
* REST APIs for rule management and request validation
* PostgreSQL for persistent rule storage
* Redis for high-performance rate limiting
* Automatic rule caching
* Prometheus metrics
* Grafana dashboards support
* Health and Readiness endpoints
* Dockerized application
* Kubernetes deployment
* Horizontal Pod Autoscaler (HPA)
* Ingress support
* Liveness and Readiness probes
* Graceful shutdown on SIGINT/SIGTERM

---

# Architecture

```
                    +----------------------+
                    |      Client          |
                    +----------+-----------+
                               |
                               |
                         HTTP REST APIs
                               |
                               v
                    +----------------------+
                    |   Go Rate Limiter    |
                    +----------+-----------+
                               |
                 +-------------+--------------+
                 |                            |
                 |                            |
                 v                            v
         PostgreSQL                     Redis
     (Rule Storage)          (Sliding Window Counters)

                 |
                 |
                 v
          Prometheus Metrics
                 |
                 v
              Grafana
```

---

# Tech Stack

| Component        | Technology              |
| ---------------- | ----------------------- |
| Language         | Go                      |
| Web Framework    | Gin                     |
| Database         | PostgreSQL              |
| Cache            | Redis                   |
| Rate Limiting    | Redis Sorted Sets + Lua |
| Metrics          | Prometheus              |
| Visualization    | Grafana                 |
| Containerization | Docker                  |
| Orchestration    | Kubernetes              |

---

# Project Structure

```
distributed-rate-limiter/

├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── config/
│   ├── database/
│   ├── handler/
│   ├── limiter/
│   ├── metrics/
│   ├── migration/
│   ├── repository/
│   └── service/
│
├── migrations/
│
├── docker/
│
├── k8s/
│
├── docs/
│
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

# Rate Limiting Algorithms

## Fixed Window

* Stores a request counter per window.
* Counter resets when the time window expires.
* Very fast.
* Can allow bursts near window boundaries.

---

## Sliding Window

Uses Redis Sorted Sets.

Each request timestamp is stored.

Workflow:

1. Remove expired timestamps.
2. Count remaining requests.
3. Compare against configured limit.
4. Add current timestamp.
5. Set expiry.

Provides smoother rate limiting than Fixed Window.

---

## Atomic Sliding Window (Lua)

The complete Sliding Window logic executes inside a Redis Lua script.

Benefits:

* Atomic execution
* No race conditions
* No distributed locking
* Better performance
* Safe with multiple application replicas

---

# Rule Management

Rules are stored in PostgreSQL.

Example:

```json
{
  "name": "Free Tier",
  "limit_count": 100,
  "window_seconds": 60
}
```

Rules are loaded by the application and used by the rate limiter.

---

# REST APIs

## Create Rule

```
POST /rules
```

Example:

```json
{
  "name": "Free Tier",
  "limit_count": 100,
  "window_seconds": 60
}
```

---

## Get Rule

```
GET /rules/{id}
```

---

## List Rules

```
GET /rules
```

---

## Check Request

```
POST /check
```

Example:

```json
{
    "key":"user:123",
    "rule_id":1
}
```

Response

```json
{
    "allowed": true
}
```

---

## Metrics

```
GET /metrics
```

Prometheus endpoint.

---

## Health

```
GET /health
```

Checks application dependencies.

---

## Ready

```
GET /ready
```

Readiness endpoint for Kubernetes.

---

# Prometheus Metrics

The application exports custom metrics including:

* Total requests
* Allowed requests
* Rejected requests
* Error count

Prometheus scrapes these metrics and Grafana visualizes them.

---

# Docker

Start dependencies

```bash
docker compose up -d
```

Build image

```bash
docker build -t distributed-rate-limiter .
```

Run

```bash
docker run -p 8080:8080 distributed-rate-limiter
```

---

# Kubernetes

The project includes Kubernetes manifests for:

* Deployment
* Service
* Ingress
* Horizontal Pod Autoscaler
* ConfigMaps
* Resource Requests & Limits
* Liveness Probe
* Readiness Probe

Deploy

```bash
kubectl apply -f k8s/
```

---

# Graceful Shutdown

The application handles:

* SIGINT
* SIGTERM

During shutdown it:

1. Stops accepting new requests.
2. Waits for in-flight requests to complete.
3. Closes Redis connections.
4. Closes PostgreSQL connections.
5. Exits cleanly.

This behavior is particularly useful during Kubernetes rolling updates and pod termination.

---

# Local Development

## Clone

```bash
git clone https://github.com/its-akshay/distributed-rate-limiter.git
cd distributed-rate-limiter
```

---

## Start dependencies

```bash
docker compose up -d postgres redis
```

---

## Run migrations

```bash
go run cmd/server/main.go
```

---

## Start application

```bash
go run cmd/server/main.go
```

---

# Example Request

Create a rule

```bash
curl -X POST http://localhost:8080/rules \
-H "Content-Type: application/json" \
-d '{
    "name":"Free Tier",
    "limit_count":100,
    "window_seconds":60
}'
```

Check a request

```bash
curl -X POST http://localhost:8080/check \
-H "Content-Type: application/json" \
-d '{
    "key":"user:123",
    "rule_id":1
}'
```

---

# Future Improvements

* Helm Charts
* GitHub Actions CI/CD
* Unit Tests
* Distributed tracing (OpenTelemetry)
* Rate limit analytics dashboard
* Dynamic configuration reload

---

# License

This project is intended for learning, experimentation, and backend engineering demonstrations.
