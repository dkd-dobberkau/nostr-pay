import { useState } from 'react'
import { QRScanner } from '../components/QRScanner'

export function PayPage() {
  const [scannedInvoice, setScannedInvoice] = useState<string | null>(null)
  const [manualInvoice, setManualInvoice] = useState('')
  const [showScanner, setShowScanner] = useState(true)

  const handleScan = (result: string) => {
    // Strip lightning: prefix if present
    const invoice = result.replace(/^lightning:/i, '')
    setScannedInvoice(invoice)
    setShowScanner(false)
  }

  const handlePay = () => {
    const invoice = scannedInvoice || manualInvoice
    if (!invoice) return

    // For MVP: open in user's Lightning wallet
    window.location.href = `lightning:${invoice}`
  }

  return (
    <div className="max-w-sm mx-auto pt-8">
      <h2 className="text-2xl font-bold mb-6">Pay</h2>

      {showScanner && !scannedInvoice ? (
        <div className="space-y-4">
          <QRScanner onScan={handleScan} />
          <div className="text-center">
            <button
              onClick={() => setShowScanner(false)}
              className="text-gray-400 hover:text-white text-sm"
            >
              Enter invoice manually
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {scannedInvoice ? (
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-4">
              <p className="text-sm text-gray-400 mb-1">Scanned Invoice</p>
              <p className="text-white font-mono text-xs break-all">
                {scannedInvoice.substring(0, 60)}...
              </p>
            </div>
          ) : (
            <div>
              <label className="block text-sm text-gray-400 mb-1">Lightning Invoice</label>
              <textarea
                value={manualInvoice}
                onChange={(e) => setManualInvoice(e.target.value)}
                placeholder="lnbc..."
                rows={4}
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white font-mono text-sm"
              />
            </div>
          )}

          <button
            onClick={handlePay}
            disabled={!scannedInvoice && !manualInvoice}
            className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-3 rounded-lg disabled:opacity-50"
          >
            Pay with Lightning Wallet
          </button>

          <button
            onClick={() => { setScannedInvoice(null); setShowScanner(true); setManualInvoice('') }}
            className="w-full text-gray-400 hover:text-white py-2"
          >
            Scan Again
          </button>
        </div>
      )}
    </div>
  )
}
