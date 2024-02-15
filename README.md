# Core SSE Demo API
#### Version 1.0

## Common

### About service
The Core SSE Demo API (here and after referred to as *API*) provides a set of methods necessary for multiple file uploading and
upload process Server Sent Events (SSE) streaming to frontend application. Can be useful for realtime data visualization, like file upload / processing progress, currency rates, etc.. It's one way, from server to client. If you need two-way communication, use websockets.

### Technologies:

Microservice core: Echo web microframework v4

Authorization: JWT token, public key verification, jwt parsing / validation

Integrations: HugginigFace Llama2 in this demo only, OpenAI, SberGigaChat and many more can be added easily

Configuration: Hocon config

Logging: LogDoc logging subsystem

Migrations: golang-migrate

Communication Bus: Asynq (Redis-based async queue) for incidents notification by telegram, sending emails, etc. 

Database: Postgres, using sqlx

Can be deployed to Docker, Dockerfile included

Observalibity: 

- Opentracing to Jaeger UI or my custom trace collector with LogDoc trace processing
- Prometheus metrics (golang standart + custom business metrics) with Grafana visualization
- LogDoc logging visualization
- Asynq queue monitoring using asynqmon

### Middlewares, Features

User claims from JWT passing through context

Radis caching by using universal cache interface and redis implementation

Resty http integration requests with curl logging and Retryier pattern implementation

Body size limiter to 10MB

Custom middlewares for Authorization header processing, custom CORS processing, multipart body validation

Rate limiter middleware, rate limit: 20 rps/sec, burst: 20 (maximum number of requests to pass at the same moment)

Teler WAF (Intrusion Detection Middleware) https://github.com/kitabisa/teler-waf.git

LogDoc logging subsystem, ClickHouse-based high performance logging collector https://logdoc.org/en/

Office, PDF, CSV uploaded content pre-processing for using with AI

pprof profiling in debug mode

SIGHUP signal config reloading

Graceful shutdown

### Building

Using Makefile:  make rebuild, restart, run, etc

### TODO

