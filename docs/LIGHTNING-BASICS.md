# Lightning Network Basics

A quick primer on the Lightning Network — what it is, how it works, and why nostr-pay uses it.

## What Is the Lightning Network?

The Lightning Network is a payment layer on top of Bitcoin. It enables fast, cheap transactions by moving most activity off the main blockchain (off-chain) while still being secured by it.

- **Bitcoin on-chain**: ~10 minute confirmation, higher fees, permanent record
- **Lightning off-chain**: instant settlement, near-zero fees, no blockchain bloat

Think of it like a bar tab: you open a tab (channel), make multiple transactions, and settle the total on-chain when you're done.

## How It Works

### Payment Channels

Two parties lock Bitcoin into a shared address (a channel). They can then send sats back and forth instantly by updating the channel balance — no miners involved.

```
Alice ──── Channel (0.01 BTC) ──── Bob
  0.006 BTC                    0.004 BTC
```

Alice sends 0.001 BTC to Bob? The balances update locally:

```
Alice ──── Channel (0.01 BTC) ──── Bob
  0.005 BTC                    0.005 BTC
```

No on-chain transaction needed. Only when the channel closes does the final balance settle on-chain.

### Routing

You don't need a direct channel to everyone. Payments route through the network:

```
Alice → Carol → Dave → Bob
```

Each hop forwards the payment. This is secured by Hash Time-Locked Contracts (HTLCs) — either the full payment goes through, or nobody loses money. There's no trust required between intermediaries.

### Invoices

To receive a payment, you create an **invoice** — a one-time payment request encoded as a string starting with `lnbc`:

```
lnbc10u1pj9...  (encodes: 1000 sats, expiry, payment hash, destination)
```

The sender scans the invoice (usually as a QR code), their wallet finds a route, and the payment settles in seconds. This is exactly what nostr-pay generates when you create a payment request.

## Key Concepts

| Term | Meaning |
|------|---------|
| **Sats** | Smallest Bitcoin unit. 1 BTC = 100,000,000 sats |
| **Channel** | A two-party payment link funded with Bitcoin |
| **Invoice** | A one-time payment request with amount, expiry, and routing info |
| **Inbound liquidity** | How many sats others can send to you (their side of the channel) |
| **Outbound liquidity** | How many sats you can send (your side of the channel) |
| **BOLT11** | The invoice format standard (what `lnbc...` strings follow) |
| **HTLC** | Hash Time-Locked Contract — the cryptographic mechanism securing multi-hop payments |

## Liquidity: The One Thing That Confuses Everyone

On-chain Bitcoin: if you have 0.1 BTC, you can receive any amount (up to block size limits).

Lightning is different. You need **inbound liquidity** to receive payments — meaning someone else has to have sats on their side of a channel to you.

```
You ──── Channel (0.01 BTC) ──── Peer
  0.01 BTC (outbound)       0.00 BTC (inbound)
```

In this scenario you can *send* 0.01 BTC but *receive* nothing. For a merchant, this matters: you need channels where the remote side has funds.

**Solutions:**
- Buy inbound liquidity from services like [Magma](https://amboss.space/magma)
- Use a wallet like Phoenix that manages channels automatically
- Spend sats first — that moves balance to the other side, creating inbound capacity
- Use a hosted node provider (Voltage, Alby Hub) that handles this for you

## How nostr-pay Uses Lightning

```
Customer's Wallet                 nostr-pay                    Your Node
     │                               │                            │
     │                    ┌───────────┴───────────┐                │
     │                    │ 1. Merchant creates   │                │
     │                    │    invoice (1000 sats) │                │
     │                    └───────────┬───────────┘                │
     │                               │──── Create Invoice ────────▶│
     │                               │◀─── BOLT11 invoice ────────│
     │◀──── QR Code (BOLT11) ────────│                            │
     │                               │                            │
     │────── Pay invoice ────────────────────────────────────────▶│
     │                               │                            │
     │                               │◀─── Payment settled ───────│
     │                               │                            │
     │                    ┌───────────┴───────────┐                │
     │                    │ 2. POS screen shows   │                │
     │                    │    green checkmark     │                │
     │                    └───────────────────────┘                │
```

1. The merchant creates an invoice via the nostr-pay UI
2. nostr-pay asks LNbits to generate a BOLT11 invoice on your Lightning node
3. The customer scans the QR code with any Lightning wallet
4. The payment routes through the Lightning Network to your node
5. LNbits detects the settled payment, nostr-pay updates the UI via WebSocket

The sats end up on your Lightning node. nostr-pay never touches the money.

## Wallets to Get Started

### For customers (paying)

Any Lightning wallet works. Install one, buy or receive some sats, scan a QR code, done.

| Wallet | Type | Best for |
|--------|------|----------|
| [Phoenix](https://phoenix.acinq.co/) | Self-custodial | Daily use, recommended |
| [Wallet of Satoshi](https://www.walletofsatoshi.com/) | Custodial | Absolute beginners |
| [Zeus](https://zeusln.com/) | Self-custodial | Node operators |
| [BlueWallet](https://bluewallet.io/) | Both options | Flexibility |

### For merchants (receiving)

You need a Lightning node behind LNbits. See [PRODUCTION.md](PRODUCTION.md) for setup options.

## Further Reading

- [Mastering the Lightning Network](https://github.com/lnbook/lnbook) — free book, best technical introduction
- [Lightning Network Spec (BOLTs)](https://github.com/lightning/bolts) — the protocol in detail
- [LNbits Documentation](https://docs.lnbits.org/) — the wallet layer nostr-pay uses
- [lightning.network](https://lightning.network/) — official site with whitepaper
