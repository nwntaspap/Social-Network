# Architectural Discussion Conclusions: Broker Swappability and Go Interface Design

This document summarizes the architectural conclusions and design patterns discussed regarding:

1. The interchangeability and hybrid use of RabbitMQ and Kafka.
2. Go-idiomatic interface design (monolithic repository vs. consumer-defined interfaces).
3. Patterns for decoupled cross-slice communication within Go vertical feature slices.

---

## 1. RabbitMQ vs. Kafka: Deep-Dive and Hybrid Strategy

While both systems handle asynchronous message transport, they are architecturally distinct and suited for different patterns. They are not direct drop-in replacements for one another without application-level adaptation.

### Key Architectural Differences

| Feature / Metric  | RabbitMQ                                                                                                  | Apache Kafka                                                                                         |
| :---------------- | :-------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------- |
| **Paradigm**      | **Smart Broker / Dumb Consumer**: Handles message routing, filtering, delivery state, and queue tracking. | **Dumb Broker / Smart Consumer**: Appends to a partitioned log; the consumer manages its own offset. |
| **Persistence**   | Ephemeral by default (messages deleted post-acknowledgement).                                             | Durable and immutable (messages retained on disk by time/size policy).                               |
| **Replayability** | No message replay.                                                                                        | High replayability (consumers can rewind offsets).                                                   |
| **Routing**       | Highly flexible routing topologies via exchanges (topic, direct, fanout, headers).                        | Simple partition-based subscription; complex routing is handled by stream processors.                |
| **Throughput**    | High (tens of thousands of messages/sec).                                                                 | Extreme (millions of messages/sec via sequential disk access & zero-copy).                           |

### The Hybrid Blueprint

For enterprise applications (like a social network), a hybrid approach leverages the strengths of both:

- **Kafka for the Analytics & Event Log Backbone**:
  - **Use Cases**: User activity tracking (clicks, views), audit logging, and data pipeline feeding.
  - **Why**: High-throughput telemetry data can be consumed by multiple analytical services (recommendation engines, search indexers) at different paces without duplicating message queues.
- **RabbitMQ for Transactional Microservice Messaging**:
  - **Use Cases**: Ephemeral tasks like notification dispatching, processing image uploads, and real-time chat routing.
  - **Why**: Easy setup for dead-letter queues (DLQs), built-in retry mechanics, and dynamic routing to individual user queues.

---

## 2. Idiomatic Go: Consuming Interfaces Where Defined

Go's proverb **"Accept interfaces, return structs"** implies that interfaces should belong to the consumer package (the package that needs the dependency) rather than the producer package (the package that implements it).

### Comparison: Repository-Level Interfaces vs. Consumer-Defined Interfaces

```
Domain-Centric Repository Interface (Proposed in Docs)
─────────────────────────────────────────────────────────────────
[ internal/user/user.go ]
  └── type Repository interface { GetByID(), Save(), Delete(), ... } (Fat)
         ▲                                   ▲
         │ (imports)                         │ (implements)
[ internal/user/commands.go ]        [ internal/user/store/sqlite.go ]


Consumer-Defined Interface (Pure Go Idiom)
─────────────────────────────────────────────────────────────────
[ internal/user/commands.go ]
  └── type UserSaver interface { Save() } (Narrow)
         ▲
         │ (implicitly implements via duck-typing)
[ internal/user/store/sqlite.go ]  <── Zero interface definitions
```

### Key Differences & Trade-offs

| Criterion                       | Unified Repository Interface (As Proposed in Docs)       | Consumer-Defined Interfaces (Refined Idiom)                               |
| :------------------------------ | :------------------------------------------------------- | :------------------------------------------------------------------------ |
| **Definition Location**         | Root entity file (e.g. `user/user.go`).                  | Consuming files (e.g. `commands.go`, `queries.go`).                       |
| **Granularity**                 | **Monolithic**: Bundles all read/write database actions. | **Granular**: Segregated into tiny contracts (`UserSaver`, `UserFinder`). |
| **Interface Segregation (ISP)** | Low. Queries depend on write methods and vice-versa.     | High. Each component only knows what it directly consumes.                |
| **Mocking / Testing**           | Requires mocking the entire database repository.         | Trivial. Mocking a single-method interface is simple.                     |
| **Circular Import Risk**        | Low (if all feature logic is co-located in one package). | Zero (packages have no upward database compile dependencies).             |

### Actionable Implementation Design

We can align the vertical slice layout proposed in `docs/plan/architecture/arch-optimized-plan.md` with Go idioms by doing the following:

1. Keep the directory tree but **omit a global interface** inside the feature root (e.g., `user/user.go`).
2. Write the database store (`store/sqlite.go`) as a concrete struct with no explicit interface declarations.
3. In `commands.go`, define a local, single-method interface (e.g., `UserSaver`).
4. In `queries.go`, define a local, read-only interface (e.g., `UserFinder`).
5. Wire the concrete store struct directly into the command/query constructors at the composition root (`bootstrap/bootstrap.go`).

---

## 3. Decoupling Cross-Slice/Module Communication

When building vertical slices, modules will inevitably need to talk to one another (e.g., `comment` needs to display user details, `chat` needs to verify group membership). To prevent circular imports and tight package coupling, we implement four decoupling strategies:

### Strategy A: ID-Only References (Data Level)

Slices **never** reference domain structs of sibling slices. They refer to them strictly by primitive IDs.

- **Yes**: `Comment` struct contains `AuthorID string`.
- **No**: `Comment` struct embeds `user.User`.
- **Data Joining**: When returning API responses, the HTTP transport layer fetches the comments, collects the unique `AuthorID`s, queries the `user` service to fetch profiles, and maps them to a DTO (Data Transfer Object) in memory.

### Strategy B: Inversion of Control / Local Interfaces (Behavior Level)

If Slice A needs a synchronous rule check from Slice B, Slice A defines a local interface. Slice B's service implements it, and they are wired at boot-time.

- **Example**: The `chat` slice needs to check if a user belongs to a group before sending a message.
- **Solution**: `chat/commands.go` declares a local interface `GroupMembershipChecker { IsMember(...) }`. The `group` service implements `IsMember`. `bootstrap.go` injects the `group` service into the `chat` service. The `chat` package **never imports** the `group` package.

### Strategy C: Asynchronous Messaging / Event Bus (Mutation Level)

If a write operation in Slice A triggers a side-effect in Slice B, Slice A publishes a domain event.

- **Example**: Creating a follow request in `follow` triggers a notification in `notification`.
- **Solution**: `follow/commands.go` publishes `follow.requested` to `platform/eventbus`. The `notification` slice runs a background worker subscribing to `follow.requested` and writes the notification to its own store. `follow` has zero compile-time awareness of `notification`.
