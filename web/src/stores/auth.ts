import { create } from 'zustand'
import { finalizeEvent, getPublicKey, type Event } from 'nostr-tools/pure'
import { hexToBytes } from '@noble/hashes/utils'

interface AuthState {
  pubkey: string | null
  secretKey: Uint8Array | null
  isLoggedIn: boolean
  login: (nsec: string) => void
  logout: () => void
  createAuthToken: (url: string, method: string) => string
}

export const useAuth = create<AuthState>((set, get) => ({
  pubkey: null,
  secretKey: null,
  isLoggedIn: false,

  login: (secretKeyHex: string) => {
    const secretKey = hexToBytes(secretKeyHex)
    const pubkey = getPublicKey(secretKey)

    set({ pubkey, secretKey, isLoggedIn: true })
  },

  logout: () => {
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
