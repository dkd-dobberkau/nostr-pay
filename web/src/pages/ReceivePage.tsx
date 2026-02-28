import { useState } from 'react'
import { QRGenerator } from '../components/QRGenerator'
import { api } from '../lib/api'
import { useAuth } from '../stores/auth'

export function ReceivePage() {
  const [amount, setAmount] = useState('')
  const [memo, setMemo] = useState('')
  const [invoice, setInvoice] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { createAuthToken } = useAuth()

  const handleGenerate = async () => {
    const sats = parseInt(amount)
    if (isNaN(sats) || sats <= 0) {
      setError('Enter a valid amount')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const url = `${window.location.origin}/api/payments/invoice`
      const token = createAuthToken(url, 'POST')
      const result = await api.createInvoice(sats, memo, token)
      setInvoice(result.bolt11)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create invoice')
    } finally {
      setLoading(false)
    }
  }

  if (invoice) {
    return (
      <div className="flex flex-col items-center gap-6 pt-8">
        <h2 className="text-2xl font-bold">{amount} sats</h2>
        <QRGenerator
          value={`lightning:${invoice}`}
          size={280}
          label="Scan to pay with Lightning"
        />
        <button
          onClick={() => { setInvoice(null); setAmount(''); setMemo('') }}
          className="text-gray-400 hover:text-white"
        >
          New Invoice
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-sm mx-auto pt-8">
      <h2 className="text-2xl font-bold mb-6">Receive Payment</h2>

      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Amount (sats)</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="1000"
            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white text-lg"
          />
        </div>

        <div>
          <label className="block text-sm text-gray-400 mb-1">Memo (optional)</label>
          <input
            type="text"
            value={memo}
            onChange={(e) => setMemo(e.target.value)}
            placeholder="What's this for?"
            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white"
          />
        </div>

        {error && <p className="text-red-400 text-sm">{error}</p>}

        <button
          onClick={handleGenerate}
          disabled={loading}
          className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-3 rounded-lg disabled:opacity-50"
        >
          {loading ? 'Generating...' : 'Generate QR Code'}
        </button>
      </div>
    </div>
  )
}
