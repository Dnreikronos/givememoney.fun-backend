# givememoney.fun — Backend

Backend API for **givememoney.fun**, a cryptocurrency donation platform for streamers. Streamers can receive donations entirely in crypto (e.g. via MetaMask, Phantom) and get real-time alerts in OBS via WebSocket.

## Features

- **Streamer auth**: Twitch, Kick, and email/password sign-up and login
- **Crypto wallets**: Streamers link wallet addresses (MetaMask, Phantom, etc.) to receive donations
- **Donations (transactions)**: Record and list donations by amount, tx hash, and message
- **Real-time alerts**: WebSocket endpoint for OBS to show donation alerts on stream
- **Sessions & JWT**: Session management with refresh tokens and configurable JWT
- **REST API**: Health check, CORS, security headers, rate limiting on auth

## Tech Stack

- **Go 1.25+** — API server
- **Gin** — HTTP router
- **GORM** — PostgreSQL ORM and migrations
- **JWT** — Access and refresh tokens
- **WebSocket** (Gorilla) — Live alerts for streamers
- **Zap** — Structured logging

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

- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `POSTGRES_PASSWORD`, `DB_NAME`
- **Server**: `PORT` (default `9090`), `FRONTEND_URL` (for CORS)
- **JWT**: `JWT_SECRET` (min 32 characters)
- **Twitch** (optional): `TWITCH_CLIENT_ID`, `TWITCH_CLIENT_SECRET`, `TWITCH_REDIRECT_URL`
- **Kick** (optional): `KICK_CLIENT_ID`, `KICK_CLIENT_SECRET`, `KICK_REDIRECT_URL`

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

- API: `http://localhost:9090`
- Health: `http://localhost:9090/health`
- pgAdmin: `http://localhost:8080` (if using compose)

Default port is **9090** unless overridden by `PORT`.

## API Overview

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/auth/register` | Email registration |
| POST | `/api/auth/login` | Email login |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Logout |
| GET | `/api/auth/twitch/login` | Twitch OAuth start |
| GET | `/api/auth/twitch/callback` | Twitch OAuth callback |
| GET | `/api/auth/kick/login` | Kick OAuth start |
| GET | `/api/auth/kick/callback` | Kick OAuth callback |
| POST | `/api/auth/wallet` | Create wallet (auth) |
| GET | `/api/auth/wallet/streamer/:streamer_id` | Get streamer wallets (auth) |
| GET | `/api/auth/wallet/:id` | Get wallet by ID (auth) |
| PUT | `/api/auth/wallet/:id` | Update wallet (auth) |
| DELETE | `/api/auth/wallet/:id` | Delete wallet (auth) |
| POST | `/api/transaction` | Create donation/transaction |
| POST | `/api/transaction/wallet/:wallet_id` | Create transaction for wallet |
| GET | `/api/transaction/wallet/:address_to_id` | List by wallet |
| GET | `/api/transaction/streamer/:streamer_id` | List by streamer |
| GET | `/api/transaction/:id` | Get transaction by ID |
| GET | `/api/transaction` | List transactions |
| GET | `/api/ws/alerts/:streamer_id` | WebSocket alerts for OBS |
| GET | `/api/alerts/:streamer_id` | Alert page for streamer |

Protected routes require a valid JWT in the `Authorization` header.

## Project Structure

```
.
├── cmd/main.go              # Entrypoint
├── internal/
│   ├── app/                 # Router, DI container
│   ├── config/              # Config loading
│   ├── controller/          # HTTP handlers (auth, wallet, transaction, websocket, alerts)
│   ├── database/connection/ # PostgreSQL + GORM migrations
│   ├── dto/                 # Request/response DTOs
│   ├── errors/              # Error types and helpers
│   ├── interfaces/          # Repository interfaces
│   ├── middleware/          # Auth, CORS, rate limit, security
│   ├── model/               # Domain models (Streamer, Wallet, Transaction, Session, etc.)
│   ├── repository/          # Data access
│   ├── service/             # Business logic (auth, JWT, Twitch/Kick providers, wallet, transaction)
│   ├── utils/               # Constants, providers (Twitch/Kick/Email, MetaMask/Phantom)
│   └── websocket/           # WebSocket hub and clients
├── docker-compose.yml       # App + Postgres + pgAdmin
├── Dockerfile               # Multi-stage Go build
├── Makefile                 # build, run, test, docker-up
├── .env.example             # Env template
└── README.md
```

## Makefile

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `givememoney.fun-backend` |
| `make run` | Run with `go run cmd/main.go` |
| `make test` | Run tests: `go test ./...` |
| `make docker-up` | Start stack with `docker-compose up -d` |

## Testing

```bash
make test
# or
go test ./...
```

## License

See repository license file.
