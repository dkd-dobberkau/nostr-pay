# nostr-pay: Instant Payment Platform Design

**Date:** 2026-02-28
**Status:** Approved

## Overview

nostr-pay is a Nostr-based instant payment platform that enables both merchant point-of-sale (POS) and peer-to-peer (P2P) payments via QR codes and Lightning Network.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Payment Method | Lightning Network (BOLT11) | Instant, low fees, proven ecosystem |
| Frontend | PWA (React/TypeScript) | Cross-platform, installable, no app store |
| Identity | NIP-46 Nostr Connect | Works on mobile without browser extensions |
| Architecture | Split (Nostr in FE, Lightning in BE) | Decentralized identity + reliable payments |
| Backend | Go | Performant, good Lightning ecosystem |
| Lightning Provider | LNbits API | Simple REST API, wallet management, self-hostable |
| Database | SQLite | Simple, no extra service, sufficient for MVP |
| Auth | NIP-98 HTTP Auth | Cryptographic, no passwords, no JWT |

## Architecture

```
┌─────────────────────────────────────────────────┐
│                    PWA (React/TS)                │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐ │
│  │ QR-Code  │  │  Wallet   │  │   Merchant    │ │
│  │ Scanner/ │  │  View     │  │   Dashboard   │ │
│  │ Generator│  │           │  │               │ │
│  └──────────┘  └──────────┘  └───────────────┘ │
│                                                  │
│  ┌──────────────────┐  ┌─────────────────────┐  │
│  │  Nostr Client     │  │  Payment Client    │  │
│  │  (nostr-tools)    │  │  (REST → Go API)   │  │
│  └────────┬─────────┘  └────────┬────────────┘  │
└───────────┼──────────────────────┼───────────────┘
            │                      │
            ▼                      ▼
   ┌─────────────────┐    ┌──────────────────┐
   │  Nostr Relays    │    │   Go API Server  │
   │                  │    │                  │
   │  - NIP-46 Auth   │    │  - LNbits Client │
   │  - NIP-44 DM     │    │  - Invoice Mgmt  │
   │  - Profile (0)   │    │  - Payment Track  │
   │  - Contacts (3)  │    │  - Merchant API   │
   └──────────────────┘    │  - SQLite         │
                           └────────┬─────────┘
                                    │
                                    ▼
                           ┌──────────────────┐
                           │     LNbits        │
                           │  - Invoice Create │
                           │  - Payment Check  │
                           │  - Webhook Notify │
                           └──────────────────┘
```

**Responsibility Split:**
- **PWA:** Nostr identity (NIP-46), QR code handling, UI, direct relay communication
- **Go API:** LNbits integration, payment tracking, merchant analytics, payment history
- **Nostr Relays:** Identity, encrypted messages, contacts
- **LNbits:** Lightning invoice creation and verification

## Nostr Integration (NIPs)

| NIP | Purpose | Details |
|-----|---------|---------|
| NIP-01 | Events & Profiles | Base events, profile metadata (Kind 0) |
| NIP-46 | Nostr Connect | Remote signing — login without browser extension |
| NIP-44 | Encrypted DMs | Encrypted receipts and payment notifications |
| NIP-47 | Wallet Connect (optional) | For users who want to use their own LN wallet |
| NIP-57 | Zaps (optional) | Standard zap compatibility for public payments |
| NIP-98 | HTTP Auth | Signed events in Authorization header for API auth |

### Custom Nostr Event Types

```
Kind 21001: Payment Request (encrypted via NIP-44)
  content: { amount_sats, memo, bolt11, merchant_npub }
  tags: [["p", recipient_pubkey], ["e", invoice_id]]

Kind 21002: Payment Confirmation (encrypted via NIP-44)
  content: { amount_sats, payment_hash, settled_at }
  tags: [["p", sender_pubkey], ["e", original_request_event_id]]
```

### NIP-46 Login Flow

1. User clicks "Login"
2. PWA generates `nostrconnect://` URI
3. User scans QR or enters bunker URL
4. NIP-46 "connect" handshake via relay
5. User authenticated — pubkey extracted

## Payment Flows

### Merchant Payment (QR at POS)

1. Merchant enters amount in POS interface
2. Go API creates Lightning invoice via LNbits
3. QR code displayed with BOLT11 invoice
4. Customer scans QR, opens in Lightning wallet
5. Customer pays invoice
6. LNbits webhook notifies Go API → payment confirmed
7. POS shows "Paid!" confirmation
8. Encrypted receipt sent via NIP-44 DM

### P2P Payment (Person to Person)

1. Sender enters recipient npub (or scans QR)
2. Sender enters amount
3. Go API creates invoice for recipient's LNbits wallet
4. Sender pays via own Lightning wallet
5. Payment confirmed via LNbits webhook
6. Nostr DM notification sent to recipient
7. Recipient's LNbits balance updated

## Go Backend

### Project Structure

```
nostr-pay/
├── cmd/server/main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware.go
│   │   ├── handlers_payment.go
│   │   ├── handlers_merchant.go
│   │   └── handlers_health.go
│   ├── lnbits/client.go
│   ├── nostr/verify.go
│   ├── payment/
│   │   ├── service.go
│   │   └── models.go
│   ├── merchant/
│   │   ├── service.go
│   │   └── models.go
│   └── store/
│       ├── sqlite.go
│       └── migrations/
├── web/                  # PWA
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

### REST API

```
GET  /api/health                  # Health check (DB + LNbits)
POST /api/auth/verify             # Verify Nostr event signature
POST /api/payments/invoice        # Create Lightning invoice
GET  /api/payments/:id            # Get payment status
GET  /api/payments/history        # Payment history (authenticated)
POST /api/payments/webhook        # LNbits webhook callback
POST /api/merchant/register       # Register as merchant
GET  /api/merchant/dashboard      # Dashboard data (today's stats)
GET  /api/merchant/transactions   # Merchant transactions
```

### Database Schema (SQLite)

```sql
CREATE TABLE users (
    pubkey TEXT PRIMARY KEY,
    is_merchant BOOLEAN DEFAULT FALSE,
    lnbits_wallet_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    bolt11 TEXT NOT NULL,
    amount_sats INTEGER NOT NULL,
    memo TEXT,
    sender_pubkey TEXT,
    receiver_pubkey TEXT NOT NULL,
    payment_hash TEXT UNIQUE,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    settled_at TIMESTAMP
);

CREATE TABLE merchant_daily_stats (
    pubkey TEXT,
    date TEXT,
    total_sats INTEGER DEFAULT 0,
    transaction_count INTEGER DEFAULT 0,
    PRIMARY KEY (pubkey, date)
);
```

## PWA Frontend

### Tech Stack

| Technology | Purpose |
|-----------|---------|
| React 19 + TypeScript | UI framework |
| Vite | Build tool |
| nostr-tools | Nostr protocol (events, NIPs, relay communication) |
| qr-scanner / qrcode.react | QR code read and generate |
| TailwindCSS | Styling |
| Workbox | PWA service worker (offline, install) |
| zustand | Lightweight state management |

### Routes

```
/                → Landing / Login (NIP-46)
/pay             → QR Scanner → Execute payment
/receive         → Enter amount → Generate QR code
/send            → Enter npub → Amount → Pay
/history         → Payment history
/merchant        → Merchant dashboard
/merchant/pos    → Point-of-Sale mode (fullscreen)
/settings        → Relay config, wallet settings
```

### POS Mode (Merchant)

Fullscreen-optimized view:
- Large numpad for amount entry
- Single-tap QR generation
- Real-time payment status (WebSocket from Go API)
- Success animation on payment confirmation
- Auto-reset for next customer

### QR Code Formats

- Lightning Invoice: `lightning:lnbc50u1p...` (BOLT11)
- Nostr npub: `nostr:npub1...` (for P2P recipient identification)
- Combo QR: `nostrpay:invoice=lnbc...&npub=npub1...&memo=Kaffee`

## Deployment

```yaml
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - LNBITS_URL=https://lnbits.example.com
      - NOSTR_RELAYS=wss://relay.damus.io,wss://nos.lol
    env_file:
      - .env
    volumes:
      - ./data:/app/data
    restart: unless-stopped

  web:
    build: ./web
    ports:
      - "3000:80"
    depends_on:
      - api
    restart: unless-stopped
```

## Security

| Area | Measure |
|------|---------|
| Auth | NIP-98 signed events — no password, no JWT, cryptographically secure |
| API | Rate limiting (per pubkey), CORS restricted to own domain |
| LNbits | API key only in backend, never exposed to frontend |
| DMs | NIP-44 encrypted — only sender/receiver can read |
| DB | Parameterized queries, no raw SQL |
| HTTPS | Required for production (reverse proxy / Caddy) |
| QR | Strict parsing — only accept known URI schemes |

## Monitoring

- `/api/health` checks DB connectivity + LNbits reachability
- Docker health checks on all services
- Structured JSON logging in Go backend
