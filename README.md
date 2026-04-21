## Bloodconnect (backend + infra)

This repository contains a Go backend API plus a background job worker (currently started inside the API process), backed by Postgres and RabbitMQ.

### Architecture (current)

```mermaid
flowchart TD
  client[Client] --> api[HTTP_API_(Go)]
  api --> db[(Postgres)]
  api --> mq[RabbitMQ]
  mq --> worker[JobWorker_(Go)]
  worker --> db
  worker --> sender[NotificationSender_(dummy)]
```

### Known TODOs / follow-ups

- **Config**: load config from env (`DATABASE_URL`, `RABBITMQ_URL`, `JWT_SECRET`, worker settings) instead of relying on `DefaultAppConfig()` hardcoded values.
- **RabbitMQ reliability**:
  - switch consumption to manual ack + retries + DLQ (avoid message loss on worker crashes)
  - add QoS/prefetch and connection/channel reconnection handling
  - add publish confirms / unroutable message handling
- **Process model**: make the worker optional (API-only vs worker-only) and add graceful shutdown via signal handling.
- **HTTP**:
  - standardize error mapping across all handlers (some endpoints still return ad-hoc error shapes/status codes)
  - decide which endpoints are truly public vs protected and enforce it consistently
- **Request eligibility edge case**: when users accept a request and then decline it, last donation date / eligibility tracking may be inconsistent and can prevent accepting a different request.
