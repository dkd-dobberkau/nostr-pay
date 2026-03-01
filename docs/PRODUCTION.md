# Production Setup

This guide explains how to run nostr-pay with real Lightning payments.

## Overview

The default docker-compose setup uses **FakeWallet** — a simulated Lightning backend for development. No real sats flow. To accept real payments, you need to connect LNbits to a real Lightning node.

```
Customer → Lightning Invoice → LNbits → Your Lightning Node → Your sats
```

## Step 1: Get a Lightning Node

You need a Lightning node that LNbits can connect to. Options from easiest to most sovereign:

| Option | Effort | Custody |
|--------|--------|---------|
| [Alby Hub](https://albyhub.com) | Low — hosted, browser-based | Semi-custodial |
| [Voltage](https://voltage.cloud) | Low — hosted LND node | You hold the keys |
| [Start9](https://start9.com) | Medium — plug-and-play home server | Fully self-hosted |
| [Umbrel](https://umbrel.com) | Medium — Raspberry Pi or old PC | Fully self-hosted |
| [LND](https://github.com/lightningnetwork/lnd) / [CLN](https://github.com/ElementsProject/lightning) | High — manual install | Fully self-hosted |

For getting started quickly, Alby Hub or Voltage are the simplest paths.

## Step 2: Configure LNbits

Edit `docker-compose.yml` and change the LNbits environment variables.

### Example: LND (REST API)

```yaml
lnbits:
  image: lnbits/lnbits:latest
  environment:
    - LNBITS_ADMIN_UI=true
    - LNBITS_BACKEND_WALLET_CLASS=LndRestWallet
    - LND_REST_ENDPOINT=https://your-lnd-host:8080
    - LND_REST_CERT=/app/data/tls.cert
    - LND_REST_MACAROON=your-admin-macaroon-hex
  volumes:
    - lnbits-data:/app/data
    - ./tls.cert:/app/data/tls.cert:ro  # mount your LND TLS cert
```

### Example: Core Lightning (CLN)

```yaml
lnbits:
  image: lnbits/lnbits:latest
  environment:
    - LNBITS_ADMIN_UI=true
    - LNBITS_BACKEND_WALLET_CLASS=CoreLightningWallet
    - CORELIGHTNING_RPC=/app/data/lightning-rpc
  volumes:
    - lnbits-data:/app/data
    - /path/to/.lightning/bitcoin/lightning-rpc:/app/data/lightning-rpc:ro
```

### Example: Alby (NWC)

```yaml
lnbits:
  image: lnbits/lnbits:latest
  environment:
    - LNBITS_ADMIN_UI=true
    - LNBITS_BACKEND_WALLET_CLASS=NWCWallet
    - NWC_PAIRING_URL=nostr+walletconnect://your-nwc-url
```

## Step 3: Create a Wallet in LNbits

1. Start the stack: `docker compose up --build -d`
2. Open LNbits at http://localhost:5001
3. Create a new wallet
4. Copy the **Admin Key** and **Invoice Key** from the wallet's API info
5. Paste them into your `.env`:

```
LNBITS_ADMIN_KEY=your-real-admin-key
LNBITS_INVOICE_KEY=your-real-invoice-key
```

6. Restart the API: `docker compose restart api`

## Step 4: Secure the Deployment

For production, you should:

- **Use HTTPS** — put nginx or Caddy in front with a TLS certificate
- **Restrict CORS** — set `CORS_ORIGINS` to your actual domain
- **Protect LNbits** — don't expose port 5001 publicly, keep it internal
- **Back up SQLite** — the database lives in `./data/nostr-pay.db`
- **Fund your node** — open Lightning channels so you have inbound liquidity to receive payments

### Example: Expose only the web frontend

```yaml
web:
  ports:
    - "443:80"  # put behind a reverse proxy with TLS

api:
  ports: []  # no direct access, only via nginx proxy

lnbits:
  ports: []  # internal only
```

## Where Are My Sats?

The money always sits on **your Lightning node**. nostr-pay and LNbits are just the interface layer.

- **View balance** — check your Lightning node or LNbits dashboard
- **Withdraw on-chain** — use your node's wallet to send to a Bitcoin address
- **Spend via Lightning** — pay invoices directly from your node

LNbits can also manage multiple wallets on top of a single node, useful if you want to separate funds per merchant or purpose.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Invoices fail to create | Check LNbits logs: `docker compose logs lnbits` |
| Payments not detected | Verify your node is online and has channels |
| "Not logged in" error | Click Login in the header and enter your Nostr key |
| LNbits can't connect to node | Check endpoint URL, certificates, and macaroon |
