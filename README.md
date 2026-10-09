# imgpipe

A decoupled, event-driven image processing service written in Go, using MongoDB for status tracking, RabbitMQ for asynchronous workload dispatch and Go's image manipulation.

## Getting Started

For local setup instructions, Docker initialization, and testing guides, see [SETUP.md](./SETUP.md). **(IN PROGRESS)**


## Roadmap
```
Current State (Stage 1)         Stage 2: Local Kubernetes           Stage 3: Cloud Primitives (AWS emulation)
┌───────────────────────┐       ┌─────────────────────────────┐     ┌─────────────────────────────────────────┐
│                       │       │ KUBERNETES CLUSTER (Kind)   │     │ KUBERNETES CLUSTER                      │
│  [ API ]   [ Worker ] │       │                             │     │                                         │
│     │          │      │       │  [ API Pods ] [Worker Pods] │     │  [ API Pods ]      [ Worker Pods ]      │
│  [RabbitMQ] [Mongo]   │  ──►  │       │             │       │ ──► │       │                   │             │
│                       │       │  [ RabbitMQ ]   [ Mongo ]   │     │  [ SQS Queue ]    [ S3 Buckets ]        │
│  (Docker Compose)     │       │                             │     │     (AWS)               (AWS)           │
└───────────────────────┘       └─────────────────────────────┘     └─────────────────────────────────────────┘
```

## Architecture & Design Decisions

`imgpipe` decouples the HTTP ingestion API from CPU-bound image transformations using a producer-consumer pattern.

```
[ Client ]
│
├─► POST /api/v1/process
│      │
│      ▼
│   [ API Service ] ──── (Insert PENDING) ────► [ MongoDB ]
│      │
│      └──────── (Publish Job JSON) ──────────► [ RabbitMQ: image_jobs ]
│                                                    │
│                                                    ▼
└─► GET /api/v1/jobs?id=... ◄── (Query Status) ── [ Worker Service ]
│
├─► Download & Resize Image
├─► Save to Storage (/renders)
├─► Update Status (COMPLETED/FAILED)
└─► Manual ACK/NACK
```

### Key Architectural Choices:
- **Asynchronous Decoupling:** The API service responds immediately with `202 Accepted` and a unique Job ID, keeping client request latency low regardless of image processing size.
- **Manual Message Acknowledgments (`ACK`/`NACK`):** RabbitMQ consumers are configured with `autoAck = false`. Messages are explicitly `ACK`ed only after successful processing and database status update. On non-recoverable errors (e.g., corrupt payloads), messages are `NACK`ed without requeueing (`requeue = false`).
- **Fair Dispatch (`QoS` Prefetching):** Workers enforce a `PrefetchCount = 1` QoS policy on the AMQP channel. This prevents a single fast-starting worker from hoarding jobs when scaling horizontally in Kubernetes.
- **Repository Pattern:** MongoDB interactions are abstracted via a repository interface, ensuring clean separation between storage mechanisms and HTTP handlers.

---

## Tech Stack

- **Language:** Go (1.27)
- **Message Broker:** RabbitMQ (AMQP 0-9-1)
- **Database:** MongoDB
- **Image Resizing:** `golang.org/x/image/draw`
- **Containerization:** Docker / Docker Compose (in progress)

---

## Repository Structure
```
.
├── cmd/
│   ├── api/          # HTTP Ingestion Service (Producer)
│   └── worker/       # Image Processing Consumer (Worker)
├── pkg/
│   ├── db/           # MongoDB Connection & Job Repository
│   ├── models/       # Job & Status Domain Structs
│   ├── processor/    # Image Fetching, Scaling & Encoding Logic
│   └── queue/        # RabbitMQ Client, Publisher & Consumer Wrappers
└── renders/          # Local output directory for processed assets
```
