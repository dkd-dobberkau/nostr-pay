import { useEffect, useState } from 'react'
import { api, type Payment } from '../lib/api'
import { useAuth } from '../stores/auth'

export function HistoryPage() {
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(true)
  const { createAuthToken, isLoggedIn } = useAuth()

  useEffect(() => {
    if (!isLoggedIn) return

    const url = `${window.location.origin}/api/payments/history`
    const token = createAuthToken(url, 'GET')

    api.getPaymentHistory(token)
      .then(setPayments)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [isLoggedIn, createAuthToken])

  if (!isLoggedIn) {
    return (
      <div className="text-center pt-8">
        <p className="text-gray-400">Login to view payment history</p>
      </div>
    )
  }

  if (loading) {
    return <div className="text-center pt-8 text-gray-400">Loading...</div>
  }

  return (
    <div className="max-w-sm mx-auto pt-4">
      <h2 className="text-2xl font-bold mb-6">Payment History</h2>

      {payments.length === 0 ? (
        <p className="text-gray-400 text-center">No payments yet</p>
      ) : (
        <div className="space-y-3">
          {payments.map((p) => (
            <div
              key={p.ID}
              className="bg-gray-900 border border-gray-800 rounded-lg p-4"
            >
              <div className="flex justify-between items-center">
                <div>
                  <p className="font-bold">
                    {p.Status === 'paid' ? '+' : ''}{p.AmountSats} sats
                  </p>
                  {p.Memo && <p className="text-sm text-gray-400">{p.Memo}</p>}
                </div>
                <span
                  className={`text-xs px-2 py-1 rounded ${
                    p.Status === 'paid'
                      ? 'bg-green-900 text-green-400'
                      : p.Status === 'expired'
                      ? 'bg-red-900 text-red-400'
                      : 'bg-yellow-900 text-yellow-400'
                  }`}
                >
                  {p.Status}
                </span>
              </div>
              <p className="text-xs text-gray-600 mt-2">
                {new Date(p.CreatedAt).toLocaleString()}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
