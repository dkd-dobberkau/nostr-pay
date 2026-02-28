# nostr-pay

Lightning payment platform with Nostr authentication.

Accept Bitcoin Lightning payments with [NIP-98](https://github.com/nostr-protocol/nips/blob/master/98.md) HTTP auth — no accounts, no emails, just your Nostr keys.

## Architecture

```
┌─────────┐     ┌──────────┐     ┌────────┐
│ React   │────▶│ Go API   │────▶│ LNbits │
│ PWA     │ WS  │ Server   │     │        │
└─────────┘     └──────────┘     └────────┘
  :3000           :8080            :5001
```

- **Frontend** — React 19, TypeScript, Vite, TailwindCSS, installable PWA
- **Backend** — Go stdlib router, SQLite, WebSocket notifications
- **Payments** — LNbits API for Lightning invoice creation and payment tracking

## Features

- NIP-98 authentication (login with Nostr private key)
- Create Lightning invoices with QR codes
- Scan & pay Lightning invoices
- Payment history
- Merchant POS mode with numpad
- Real-time payment notifications via WebSocket
- Session-based key storage (cleared on tab close)

## Quick Start

### Prerequisites

- Docker & Docker Compose
- An LNbits instance (included via docker-compose for development)

### Run

```bash
cp .env.example .env  # configure LNBITS_API_KEY
docker compose up --build -d
```

Open http://localhost:3000, click **Login**, generate a test key, and start accepting payments.

### Services

| Service | URL | Description |
|---------|-----|-------------|
| Web | http://localhost:3000 | React PWA frontend |
| API | http://localhost:8080 | Go backend |
| LNbits | http://localhost:5001 | Lightning wallet (FakeWallet in dev) |

## Development

### Backend

```bash
go run ./cmd/server/
```

### Frontend

```bash
cd web
npm install
npm run dev
```

### API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/payments/invoice` | NIP-98 | Create Lightning invoice |
| GET | `/api/payments/:id` | NIP-98 | Get payment status |
| GET | `/api/payments/history` | NIP-98 | Payment history |
| GET | `/api/health` | — | Health check |
| GET | `/ws` | — | WebSocket notifications |

## Tech Stack

- **Go 1.24+** — stdlib net/http router, ncruces/go-sqlite3, go-nostr, gorilla/websocket
- **React 19** — TypeScript, Vite, TailwindCSS, nostr-tools, zustand, qrcode.react, html5-qrcode
- **LNbits** — Lightning wallet backend
- **Docker** — Multi-stage builds, nginx for frontend

## License

[MIT](LICENSE)
