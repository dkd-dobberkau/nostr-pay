# No Accounts, No Emails — Just Lightning Payments with Your Nostr Keys

What if accepting Bitcoin payments was as simple as logging in with your Nostr identity? No sign-up forms, no email verification, no password resets. Just cryptographic keys you already own.

That's the idea behind **nostr-pay** — a small, self-hosted payment platform that pairs Lightning Network invoices with Nostr-based authentication.

## The Problem with Payment Platforms

Most payment solutions require you to create yet another account. You hand over your email, set a password, maybe verify your phone number — all before you can generate your first invoice. For merchants who just want to accept sats at a market stall or a pop-up shop, that's a lot of friction for a simple task.

## Enter Nostr

Nostr gives us something powerful: a universal identity layer based on public-key cryptography. If you have a Nostr keypair, you already have an identity. NIP-98 extends this to HTTP — it lets you sign API requests the same way you sign Nostr events. The server verifies your signature, knows who you are, and never needs to store a password.

nostr-pay uses exactly this. You log in with your Nostr private key, and every API call is authenticated with a signed NIP-98 event. No sessions on the server, no cookies, no OAuth dance.

## What It Does

nostr-pay is a compact, self-hosted stack with three services:

- A **React PWA** that works on any phone — installable, fast, and offline-capable
- A **Go API server** that handles authentication, invoices, and payment tracking
- An **LNbits** instance that talks to the Lightning Network

As a user, you open the app, tap Login, and either paste your Nostr hex key or generate a test key for development. From there you can:

- **Receive payments** — enter an amount, get a QR code, have someone scan it
- **Pay invoices** — scan a Lightning QR code with your camera
- **Browse history** — see all your past payments
- **Use POS mode** — a full-screen numpad designed for merchants serving customers face-to-face

Payment notifications arrive in real-time over WebSocket, so the POS screen flips to a green checkmark the moment a customer pays.

## Why Self-Hosted?

There's no central server. You run your own instance, connect it to your own LNbits wallet (or any Lightning backend), and you're in control. The keys never leave your browser tab — they're stored in `sessionStorage` and wiped when you close the tab.

This matters. A payment platform that phones home to someone else's server with your credentials defeats the purpose of using Nostr in the first place.

## The Tech in Brief

The backend is plain Go with the standard library router — no frameworks, no magic. SQLite handles persistence, which means zero database infrastructure. The frontend is React 19 with TypeScript, Vite, and TailwindCSS. Everything ships in Docker containers behind an nginx reverse proxy.

The entire stack starts with one command:

```bash
docker compose up --build -d
```

## Who Is This For?

Anyone who wants to accept Lightning payments without trusting a third party:

- **Market vendors** who want a quick POS on their phone
- **Developers** exploring Nostr-based authentication patterns
- **Bitcoin enthusiasts** who prefer self-sovereign tooling

nostr-pay is intentionally small. It doesn't try to be a full merchant suite or a wallet. It's a focused tool that does one thing well: turn Nostr identities into payment endpoints.

## Try It

The project is open source under MIT. Clone it, spin it up, and generate a test key to see it in action:

→ [github.com/dkd-dobberkau/nostr-pay](https://github.com/dkd-dobberkau/nostr-pay)

If you already have a Nostr keypair, you're one click away from your first Lightning invoice.

*Note: nostr-pay is in early alpha — your mileage may vary.*
