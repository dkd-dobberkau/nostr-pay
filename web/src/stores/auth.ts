import { create } from 'zustand'
import { finalizeEvent, getPublicKey, type Event } from 'nostr-tools/pure'
import { bytesToHex } from 'nostr-tools/utils'

function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2)
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16)
  }
  return bytes
}

export { bytesToHex }

const SESSION_KEY = 'nostr-pay-sk'

function loadFromSession(): { pubkey: string; secretKey: Uint8Array } | null {
  const stored = sessionStorage.getItem(SESSION_KEY)
  if (!stored || stored.length !== 64) return null
  try {
    const secretKey = hexToBytes(stored)
    const pubkey = getPublicKey(secretKey)
    return { pubkey, secretKey }
  } catch {
    sessionStorage.removeItem(SESSION_KEY)
    return null
  }
}

interface AuthState {
  pubkey: string | null
  secretKey: Uint8Array | null
  isLoggedIn: boolean
  login: (nsec: string) => void
  logout: () => void
  createAuthToken: (url: string, method: string) => string
}

const initial = loadFromSession()

export const useAuth = create<AuthState>((set, get) => ({
  pubkey: initial?.pubkey ?? null,
  secretKey: initial?.secretKey ?? null,
  isLoggedIn: !!initial,

  login: (secretKeyHex: string) => {
    const secretKey = hexToBytes(secretKeyHex)
    const pubkey = getPublicKey(secretKey)
    sessionStorage.setItem(SESSION_KEY, secretKeyHex)
    set({ pubkey, secretKey, isLoggedIn: true })
  },

  logout: () => {
    sessionStorage.removeItem(SESSION_KEY)
    set({ pubkey: null, secretKey: null, isLoggedIn: false })
  },

  createAuthToken: (url: string, method: string): string => {
    const { secretKey } = get()
    if (!secretKey) throw new Error('Not logged in')

    const event: Event = finalizeEvent({
      kind: 27235,
      created_at: Math.floor(Date.now() / 1000),
      tags: [
        ['u', url],
        ['method', method],
      ],
      content: '',
    }, secretKey)

    const eventJson = JSON.stringify(event)
    const token = btoa(eventJson)
    return `Nostr ${token}`
  },
}))
