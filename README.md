# http-common-go

**Shared Go infrastructure library for building HTTP-based microservices.**

`http-common-go` provides standardized building blocks for logging, database access, caching, middleware, and response handling — allowing consistent, observable, and maintainable service implementations across your ecosystem.

---

## 📦 Overview

This library provides composable primitives for common microservice needs:

- 🚀 **Structured Logging** — consistent, colorized log output  
- 🧩 **Middleware** — CORS, recovery, request tracing, and logging  
- 🗄️ **Database Connectors** — PostgreSQL and MongoDB connection utilities  
- ⚡ **Cache** — Redis-based JSON caching abstraction  
- 📤 **Response Utilities** — standardized JSON envelopes and writers  

All components are designed to be imported individually and combined within your own service modules.

---

## 🧱 Package Layout

```
pkg/
├── cache/          # Redis-based caching abstraction
├── db/
│   ├── mongo/      # MongoDB connection utilities
│   └── postgres/   # PostgreSQL connection utilities
├── log/
│   ├── formatter/  # Custom Logrus formatter
│   └── logger/     # Structured logger setup and helpers
├── middleware/
│   ├── context/    # Gin context propagation helpers
│   ├── cors/       # CORS middleware
│   ├── logger/     # Request logging + OpenTelemetry traceparent support
│   └── recovery/   # Panic recovery middleware
└── response/       # Standardized API responses
```

---

## ⚙️ Installation

```bash
go get github.com/ranorsolutions/http-common-go
```

> Requires **Go 1.22+**

---

## 🪵 Logging

```go
import "github.com/ranorsolutions/http-common-go/pkg/log/logger"

func main() {
    log, _ := logger.New("user-service", "1.0.0", true)
    log.Info("Service started successfully")
}
```

### Gin Middleware Integration

```go
import (
    "github.com/gin-gonic/gin"
    logmw "github.com/ranorsolutions/http-common-go/pkg/middleware/logger"
    "github.com/ranorsolutions/http-common-go/pkg/log/logger"
)

func setupRouter() *gin.Engine {
    r := gin.New()
    appLogger, _ := logger.New("user-service", "1.0.0", true)
    r.Use(logmw.Middleware(appLogger))
    return r
}
```

Logs automatically include:
- `request_id`, `trace_id`, `span_id` (W3C traceparent support)  
- `service`, `version`, `method`, `path`, `status`, `latency`, `clientIP`

---

## 🌐 Middleware

### CORS
```go
r.Use(cors.CORSMiddleware(nil)) // uses permissive defaults
```

### Recovery
```go
r.Use(recovery.Middleware(nil)) // recovers from panics and logs error
```

### Context Propagation
```go
r.Use(context.GinContextToContextMiddleware())
// later in goroutines:
gc := context.GinContextFromContext(ctx)
```

---

## 🗄️ Database

### PostgreSQL
```go
import "github.com/ranorsolutions/http-common-go/pkg/db/postgres"

conn := postgres.GetURIFromEnv() // reads DB_* env vars
db, err := postgres.Connect(conn)
if err != nil {
    panic(err)
}
defer db.Close()
```

### MongoDB
```go
import "github.com/ranorsolutions/http-common-go/pkg/db/mongo"

cfg, _ := mongo.GetFromEnv()
uri := cfg.URI()
mongoDB, _ := mongo.New("appdb", uri)
if err := mongoDB.HealthCheck(); err != nil {
    panic(err)
}
```

---

## ⚡ Cache

```go
import (
    "time"
    "context"
    "github.com/redis/go-redis/v9"
    "github.com/ranorsolutions/http-common-go/pkg/cache"
)

func main() {
    ctx := context.Background()
    client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    c := cache.NewRedisCache(client, 10*time.Minute)

    type User struct { Name string; Age int }
    _ = c.SetJSON(ctx, "user:1", User{"Alice", 30}, 0)

    var u User
    found, _ := c.GetJSON(ctx, "user:1", &u)
    if found {
        fmt.Println("Cached user:", u.Name)
    }
}
```

---

## 📤 Response Helpers

```go
import "github.com/ranorsolutions/http-common-go/pkg/response"

func handler(w http.ResponseWriter, r *http.Request) {
    _ = response.WriteJSON(w, 200, "ok", map[string]string{"msg": "success"})
}
```

---

## 🧭 Tracing

The `logger` middleware automatically handles W3C trace context propagation (`traceparent` and `tracestate` headers).  
Every log entry includes `trace_id` and `span_id` fields, enabling correlation across distributed services.

Example response headers:
```
X-Request-ID: 5be143d8-65b7-4ac1-9b38-0e6b87dcfc41
traceparent: 00-9f5ec3d8247340d2a4460ad58bcd3c7f-21b1c917c50f3a6a-01
```

---

## 🧪 Testing

Run all tests:
```bash
go test ./pkg/... -v
```

The Redis cache uses [miniredis](https://github.com/alicebob/miniredis) for in-memory tests, no external dependencies required.

---

## 🧩 Example Service Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/ranorsolutions/http-common-go/pkg/log/logger"
    logmw "github.com/ranorsolutions/http-common-go/pkg/middleware/logger"
    "github.com/ranorsolutions/http-common-go/pkg/middleware/cors"
    "github.com/ranorsolutions/http-common-go/pkg/middleware/recovery"
)

func main() {
    r := gin.New()
    log, _ := logger.New("example-service", "1.0.0", true)

    r.Use(recovery.Middleware(nil))
    r.Use(cors.CORSMiddleware(nil))
    r.Use(logmw.Middleware(log))

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}
```

---

## 📄 License

This project is licensed under the **MIT License**.  
See the full text in [LICENSE](./LICENSE).

---

## 👩‍💻 Maintainer

**Abigail Ranson**  
Maintainer — [Ranor Solutions](https://ranorsolutions.com)  
© 2025 Abigail Ranson. All rights reserved.

