import { useState, useCallback } from 'react'
import { QRGenerator } from '../components/QRGenerator'
import { api } from '../lib/api'
import { useAuth } from '../stores/auth'

type POSState = 'input' | 'waiting' | 'paid'

export function MerchantPOS() {
  const [amount, setAmount] = useState('0')
  const [state, setState] = useState<POSState>('input')
  const [invoice, setInvoice] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { createAuthToken } = useAuth()

  const handleNumpad = useCallback((key: string) => {
    setAmount((prev) => {
      if (key === 'C') return '0'
      if (key === '←') return prev.length > 1 ? prev.slice(0, -1) : '0'
      if (prev === '0') return key
      return prev + key
    })
  }, [])

  const handleCharge = async () => {
    const sats = parseInt(amount)
    if (sats <= 0) return

    setError(null)
    setState('waiting')

    try {
      const url = `${window.location.origin}/api/payments/invoice`
      const token = createAuthToken(url, 'POST')
      const result = await api.createInvoice(sats, `POS Payment`, token)
      setInvoice(result.bolt11)

      // Poll for payment status
      const pollInterval = setInterval(async () => {
        try {
          const paymentUrl = `${window.location.origin}/api/payments/${result.payment_id}`
          const pollToken = createAuthToken(paymentUrl, 'GET')
          const payment = await api.getPayment(result.payment_id, pollToken)
          if (payment.Status === 'paid') {
            clearInterval(pollInterval)
            setState('paid')
          }
        } catch {
          // Ignore poll errors
        }
      }, 2000)

      // Timeout after 10 minutes
      setTimeout(() => {
        clearInterval(pollInterval)
      }, 600_000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
      setState('input')
    }
  }

  const handleReset = () => {
    setAmount('0')
    setState('input')
    setInvoice(null)
    setError(null)
  }

  // Paid state
  if (state === 'paid') {
    return (
      <div className="fixed inset-0 bg-green-950 flex flex-col items-center justify-center gap-6 z-50">
        <div className="text-6xl">&#10003;</div>
        <h2 className="text-4xl font-bold text-green-400">Paid!</h2>
        <p className="text-2xl text-green-300">{parseInt(amount).toLocaleString()} sats</p>
        <button
          onClick={handleReset}
          className="mt-8 bg-green-800 hover:bg-green-700 text-white font-bold py-4 px-12 rounded-xl text-xl"
        >
          Next Customer
        </button>
      </div>
    )
  }

  // QR display state
  if (state === 'waiting' && invoice) {
    return (
      <div className="fixed inset-0 bg-gray-950 flex flex-col items-center justify-center gap-6 z-50">
        <h2 className="text-3xl font-bold">{parseInt(amount).toLocaleString()} sats</h2>
        <QRGenerator value={`lightning:${invoice}`} size={300} />
        <p className="text-gray-400 animate-pulse">Waiting for payment...</p>
        <button
          onClick={handleReset}
          className="text-gray-600 hover:text-gray-400 text-sm mt-4"
        >
          Cancel
        </button>
      </div>
    )
  }

  // Numpad input state
  const numpadKeys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', 'C', '0', '←']

  return (
    <div className="fixed inset-0 bg-gray-950 flex flex-col z-50">
      <div className="flex-1 flex flex-col items-center justify-center">
        <p className="text-gray-500 text-sm mb-2">Amount (sats)</p>
        <p className="text-5xl font-bold tabular-nums">
          {parseInt(amount).toLocaleString()}
        </p>
      </div>

      {error && <p className="text-red-400 text-center text-sm mb-2">{error}</p>}

      <div className="grid grid-cols-3 gap-2 p-4 max-w-sm mx-auto w-full">
        {numpadKeys.map((key) => (
          <button
            key={key}
            onClick={() => handleNumpad(key)}
            className="bg-gray-900 hover:bg-gray-800 text-white text-2xl font-bold py-5 rounded-xl active:bg-gray-700"
          >
            {key}
          </button>
        ))}
      </div>

      <div className="p-4 max-w-sm mx-auto w-full">
        <button
          onClick={handleCharge}
          disabled={amount === '0'}
          className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-4 rounded-xl text-xl disabled:opacity-30"
        >
          Charge
        </button>
      </div>
    </div>
  )
}
