# givememoney.fun — Backend

Backend API for **givememoney.fun**, a cryptocurrency donation platform for streamers. Streamers can receive donations entirely in crypto (e.g. via MetaMask, Phantom) and get real-time alerts in OBS via WebSocket.

## Features

- **Streamer auth**: Twitch, Kick, and email/password sign-up and login
- **Crypto wallets**: Streamers link wallet addresses (MetaMask, Phantom) to receive donations in ETH, USDT, or USDC
- **Donations (transactions)**: Record and list donations by amount, tx hash, currency, and message
- **Real-time alerts**: WebSocket endpoint for OBS to show donation alerts on stream with keep-alive (ping/pong)
- **Multi-session management**: Track active sessions across devices with device type, IP, and user-agent detection
- **JWT with refresh token rotation**: 15-min access tokens, 24-hour refresh tokens with hash-based storage
- **Security**: CORS, CSP, HSTS, XSS protection headers, IP-based rate limiting, bcrypt password hashing
- **Observability**: Structured logging with Zap, request ID tracing, ELK stack integration (Elasticsearch, Logstash, Kibana, Filebeat)
- **REST API**: Health check, standardized error responses, request validation

## Tech Stack

- **Go 1.25+** — API server
- **Gin** — HTTP router
- **GORM** — PostgreSQL ORM and auto-migrations
- **JWT** (golang-jwt/jwt/v5) — Access and refresh tokens
- **WebSocket** (Gorilla) — Live alerts for streamers
- **Zap** — Structured logging
- **ELK Stack** — Elasticsearch, Logstash, Kibana, Filebeat for centralized log management
- **bcrypt** — Password hashing
- **x/time** — Token-bucket rate limiting

## Prerequisites

- Go 1.25 or later
- PostgreSQL
- (Optional) Docker and Docker Compose for running app + DB in containers

## Quick Start

### 1. Clone and install dependencies

```bash
git clone https://github.com/Dnreikronos/givememoney.fun-backend.git
cd givememoney.fun-backend
go mod download
```

### 2. Environment variables

Copy the example env and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with at least:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | — |
| `DB_NAME` | Database name | `givememoney` |
| `POSTGRES_TIME_ZONE` | Database timezone | `America/Sao_Paulo` |
| `PORT` | API server port | `9090` |
| `FRONTEND_URL` | Allowed CORS origin | `http://localhost:3000` |
| `JWT_SECRET` | Signing key (min 32 chars) | — |
| `GO_ENV` | `development` or `production` | `development` |
| `SERVICE_NAME` | Logger service name | `givememoney` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `TWITCH_CLIENT_ID` | Twitch OAuth client ID (optional) | — |
| `TWITCH_CLIENT_SECRET` | Twitch OAuth secret (optional) | — |
| `TWITCH_REDIRECT_URL` | Twitch callback URL (optional) | — |
| `KICK_CLIENT_ID` | Kick OAuth client ID (optional) | — |
| `KICK_CLIENT_SECRET` | Kick OAuth secret (optional) | — |
| `KICK_REDIRECT_URL` | Kick callback URL (optional) | — |

See [.env.example](.env.example) for full list and comments.

### 3. Run the app

**Option A — Local Go**

```bash
make run
# or
go run cmd/main.go
```

**Option B — Docker Compose (app + PostgreSQL + pgAdmin)**

```bash
make docker-up
# or
docker-compose up -d
```

| Service | URL | Description |
|---------|-----|-------------|
| API | `http://localhost:9090` | Backend server |
| Health | `http://localhost:9090/health` | Health check |
| pgAdmin | `http://localhost:8080` | Database UI |
| Kibana | `http://localhost:5601` | Log visualization (ELK) |
| Elasticsearch | `http://localhost:9200` | Log storage (ELK) |

Default port is **9090** unless overridden by `PORT`.

## API Overview

### Public

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |

### Authentication (rate-limited: 2 req/s, burst 5)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Email registration |
| POST | `/api/auth/login` | Email login |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout / invalidate session |
| GET | `/api/auth/twitch/login` | Twitch OAuth redirect |
| GET | `/api/auth/twitch/callback` | Twitch OAuth callback |
| GET | `/api/auth/kick/login` | Kick OAuth redirect (PKCE) |
| GET | `/api/auth/kick/callback` | Kick OAuth callback |

### Sessions (protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/session` | Create session |
| GET | `/api/auth/session` | Get current session |
| DELETE | `/api/auth/session` | Delete current session |
| GET | `/api/auth/sessions` | List all active sessions |

### Wallets (protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/wallet` | Create wallet |
| GET | `/api/auth/wallet/:id` | Get wallet by ID |
| GET | `/api/auth/wallet/streamer/:streamer_id` | List streamer wallets |
| GET | `/api/auth/wallet/address/:wallet_address` | Get wallet by address |
| PUT | `/api/auth/wallet/:id` | Update wallet |
| DELETE | `/api/auth/wallet/:id` | Delete wallet |

### Transactions / Donations (public)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/transaction` | Create donation |
| POST | `/api/transaction/wallet/:wallet_id` | Create donation for wallet |
| GET | `/api/transaction/:id` | Get by ID |
| GET | `/api/transaction` | List all |
| GET | `/api/transaction/wallet/:address_to_id` | List by wallet |
| GET | `/api/transaction/streamer/:streamer_id` | List by streamer |

### Real-time Alerts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/ws/alerts/:streamer_id` | WebSocket connection for OBS |
| GET | `/ws/streamer/:streamer_id` | Alternative WebSocket endpoint |
| GET | `/api/alerts/:streamer_id` | HTML alert overlay page |

Protected routes require a valid JWT in the `Authorization: Bearer <token>` header.

## Project Structure

```
.
├── cmd/main.go                  # Entrypoint
├── internal/
│   ├── app/
│   │   ├── container.go         # DI container setup
│   │   └── router.go            # Route definitions & middleware wiring
│   ├── config/                  # Config loading from .env
│   ├── controller/              # HTTP handlers
│   │   ├── auth_controller.go         # OAuth2 flows (Twitch/Kick)
│   │   ├── email_auth_controller.go   # Email/password auth
│   │   ├── session_controller.go      # Session management
│   │   ├── wallet_controller.go       # Wallet CRUD
│   │   ├── transaction_controller.go  # Donation CRUD
│   │   ├── websocket_controller.go    # WebSocket upgrade
│   │   └── alert_controller.go        # OBS alert overlay
│   ├── database/connection/     # PostgreSQL + GORM auto-migrations
│   ├── dto/                     # Request/response DTOs
│   ├── errors/                  # AppError types with HTTP status codes
│   ├── interfaces/              # Repository interfaces
│   ├── middleware/
│   │   ├── auth.go              # JWT Bearer token validation
│   │   ├── cors.go              # CORS configuration
│   │   ├── security.go          # Security headers (CSP, HSTS, X-Frame-Options)
│   │   ├── rate_limiter.go      # IP-based token-bucket rate limiting
│   │   ├── error_handler.go     # Standardized error responses
│   │   ├── request_logger.go    # Zap request logging
│   │   ├── request_id.go        # UUID request tracing
│   │   ├── password_validation.go  # Password strength rules
│   │   └── validation.go        # Request body validation
│   ├── model/                   # GORM domain models
│   │   ├── streamer.go          # Streamer (user)
│   │   ├── wallet.go            # Wallet (MetaMask/Phantom)
│   │   ├── transaction.go       # Donation/transaction
│   │   ├── session.go           # Active session tracking
│   │   └── refresh_token.go     # Hashed refresh tokens
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic
│   │   ├── auth_service.go      # OAuth provider orchestration
│   │   ├── jwt_service.go       # JWT generation & validation
│   │   ├── session_service.go   # Session lifecycle
│   │   ├── wallet_service.go    # Wallet operations
│   │   ├── transaction_service.go  # Transactions + WebSocket broadcast
│   │   ├── twitch_*.go          # Twitch OAuth provider & API
│   │   ├── kick_*.go            # Kick OAuth provider & API (PKCE)
│   │   └── logger_service.go    # Zap logger factory
│   ├── utils/                   # Constants, providers, password hashing
│   ├── validator/               # Custom validators (wallet address)
│   └── websocket/
│       ├── hub.go               # Broadcast hub (per-streamer channels)
│       └── client.go            # Read/write pumps with keep-alive
├── elk/
│   ├── filebeat/filebeat.yml    # Docker log ingestion
│   └── logstash/
│       ├── config/logstash.yml  # Logstash settings
│       └── pipeline/givememoney.conf  # Log parsing & enrichment
├── .github/
│   └── pull_request_template.md # PR template
├── docker-compose.yml           # App + Postgres + pgAdmin + ELK stack
├── Dockerfile                   # Multi-stage Go build (Alpine)
├── Makefile                     # Build, run, test, docker, ELK commands
├── .env.example                 # Environment template
└── README.md
```

## Makefile

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `givememoney.fun-backend` |
| `make run` | Run with `go run cmd/main.go` |
| `make test` | Run tests: `go test ./...` |
| `make docker-up` | Start full stack with `docker-compose up -d` |
| `make elk-up` | Start ELK stack only (Elasticsearch, Logstash, Kibana, Filebeat) |
| `make elk-down` | Stop and remove ELK containers |
| `make elk-logs` | Tail ELK container logs |
| `make elk-reset` | Stop ELK, remove containers and volumes (full reset) |

## Authentication & Security

### JWT Tokens

| Token | TTL | Storage |
|-------|-----|---------|
| Access token | 15 minutes | Client (Authorization header) |
| Refresh token | 24 hours | Database (SHA-256 hashed) |

Tokens are signed with HS256. Refresh tokens are rotated on use (old token revoked, new token issued).

### OAuth Providers

- **Twitch** — Standard OAuth2 code exchange flow
- **Kick** — OAuth2 with PKCE (Proof Key for Code Exchange)
- **Email/Password** — bcrypt-hashed passwords, minimum 8 characters

### Cookie Configuration (for OAuth callbacks)

Cookies are `HttpOnly` and `Secure` in production (`GO_ENV=production`). `SameSite` defaults to `Lax`.

### Security Headers

All responses include:
- `Content-Security-Policy`: `default-src 'self'; script-src 'self'`
- `X-Frame-Options`: `DENY`
- `X-Content-Type-Options`: `nosniff`
- `X-XSS-Protection`: `1; mode=block`
- `Strict-Transport-Security`: enabled in production (1 year max-age)

### Rate Limiting

IP-based token-bucket rate limiting with automatic cleanup of stale entries:

| Scope | Rate | Burst |
|-------|------|-------|
| Auth routes (login/register) | 2 req/s | 5 |
| Strict routes | 5 req/s | 10 |

Returns `429 Too Many Requests` with `Retry-After` header when exceeded.

## Middleware Stack

Requests pass through the following middleware chain:

1. **Request ID** — Injects a UUID for distributed tracing
2. **Request Logger** — Logs method, path, status, latency, and client IP via Zap
3. **Error Handler** — Converts `AppError` to standardized JSON responses
4. **CORS** — Allows `FRONTEND_URL` origin with credentials
5. **Security Headers** — CSP, HSTS, X-Frame-Options, etc.
6. **Rate Limiter** — Per-route IP-based limiting (auth routes only)
7. **Auth** — Per-route JWT Bearer token validation

## Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "error": "unauthorized",
  "message": "Invalid or expired token",
  "code": "unauthorized"
}
```

Error codes: `bad_request` (400), `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` (409), `validation_error` (400), `internal_error` (500).

## WebSocket Alerts

### Flow

1. Donor creates a transaction via `POST /api/transaction`
2. `TransactionService` broadcasts the donation to the WebSocket hub
3. Hub routes the message to all clients connected to the streamer's channel
4. OBS browser source displays the alert overlay

### Connection Details

- **Ping interval**: 54 seconds
- **Read deadline**: 60 seconds (connection closed if no pong received)
- **Message format**: JSON with transaction data (amount, currency, message, sender address)

Connect from OBS:
```
ws://localhost:9090/api/ws/alerts/<streamer_id>
```

### Alert Page

The built-in alert page at `/api/alerts/:streamer_id` serves an HTML overlay with embedded WebSocket client that can be used directly as an OBS browser source.

## Supported Currencies & Wallets

### Wallet Providers

| Provider | Address Format |
|----------|---------------|
| MetaMask | 64-character hex string |
| Phantom | 64-character hex string |

Addresses are normalized to lowercase and must be unique.

### Currencies

| Currency | Type |
|----------|------|
| ETH | Native |
| USDT | Stablecoin |
| USDC | Stablecoin |

Default currency is ETH if not specified.

## ELK Stack (Logging)

The project includes a full ELK stack for centralized log management:

- **Filebeat** — Collects Docker container logs and forwards to Logstash
- **Logstash** — Parses and enriches logs (drops `/health` noise, classifies HTTP status codes, converts latency to milliseconds)
- **Elasticsearch** — Stores logs in daily indices (`givememoney-logs-YYYY.MM.dd`)
- **Kibana** — Web UI for log visualization and dashboards at `http://localhost:5601`

Start the ELK stack:

```bash
make elk-up
```

## Testing

```bash
make test
# or
go test ./...
```

## 🤝 Our Team

Meet the builders of the project:

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/yuribodo" title="Yuri Bodó">
        <img src="https://avatars3.githubusercontent.com/u/83407152" width="100px;" alt="Yuri Bodó"/><br>
        <sub><b>Yuri Bodó</b></sub>
      </a>
      <br />
      <a href="https://linkedin.com/in/mario-lara-1a801b272">
        <img src="https://img.shields.io/badge/LinkedIn-0077B5?style=flat&logo=linkedin&logoColor=white" alt="LinkedIn Badge"/>
      </a>
    </td>
    <td align="center">
      <a href="https://github.com/Dnreikronos" title="João Soares">
        <img src="https://avatars3.githubusercontent.com/u/37777652" width="100px;" alt="João Soares"/><br>
        <sub><b>João Soares</b></sub>
      </a>
      <br />
      <a href="https://linkedin.com/in/joao-roberto-lawall-soares-a58468242">
        <img src="https://img.shields.io/badge/LinkedIn-0077B5?style=flat&logo=linkedin&logoColor=white" alt="LinkedIn Badge"/>
      </a>
    </td>
  </tr>
</table>


## License

See repository license file.
