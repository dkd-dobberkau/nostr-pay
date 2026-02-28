import { useState } from 'react'
import { generateSecretKey } from 'nostr-tools/pure'
import { bytesToHex } from 'nostr-tools/utils'
import { useAuth } from '../stores/auth'

interface LoginModalProps {
  onClose: () => void
}

export function LoginModal({ onClose }: LoginModalProps) {
  const [keyInput, setKeyInput] = useState('')
  const [error, setError] = useState<string | null>(null)
  const { login } = useAuth()

  const handleLogin = () => {
    const key = keyInput.trim()
    if (!/^[0-9a-f]{64}$/.test(key)) {
      setError('Invalid key — enter 64 hex characters')
      return
    }
    try {
      login(key)
      onClose()
    } catch {
      setError('Invalid private key')
    }
  }

  const handleGenerate = () => {
    const sk = generateSecretKey()
    setKeyInput(bytesToHex(sk))
    setError(null)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
      <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-full max-w-sm mx-4">
        <h2 className="text-xl font-bold mb-4">Login</h2>

        <label className="block text-sm text-gray-400 mb-1">
          Private Key (hex)
        </label>
        <input
          type="password"
          value={keyInput}
          onChange={(e) => { setKeyInput(e.target.value); setError(null) }}
          placeholder="64-character hex secret key"
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white font-mono text-sm mb-2"
          autoFocus
        />

        {error && <p className="text-red-400 text-sm mb-2">{error}</p>}

        <button
          onClick={handleGenerate}
          className="text-sm text-amber-400 hover:text-amber-300 mb-4 block"
        >
          Generate Test Key
        </button>

        <div className="flex gap-3">
          <button
            onClick={onClose}
            className="flex-1 border border-gray-700 text-gray-300 hover:bg-gray-800 font-bold py-3 rounded-lg"
          >
            Cancel
          </button>
          <button
            onClick={handleLogin}
            className="flex-1 bg-amber-500 hover:bg-amber-600 text-black font-bold py-3 rounded-lg"
          >
            Login
          </button>
        </div>
      </div>
    </div>
  )
}
