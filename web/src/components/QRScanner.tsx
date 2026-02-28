import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode } from 'html5-qrcode'

interface QRScannerProps {
  onScan: (result: string) => void
  onError?: (error: string) => void
}

export function QRScanner({ onScan, onError }: QRScannerProps) {
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const [isScanning, setIsScanning] = useState(false)
  const containerRef = useRef<string>(`qr-reader-${Date.now()}`)

  useEffect(() => {
    const scanner = new Html5Qrcode(containerRef.current)
    scannerRef.current = scanner

    scanner.start(
      { facingMode: 'environment' },
      { fps: 10, qrbox: { width: 250, height: 250 } },
      (decodedText) => {
        scanner.stop().then(() => {
          setIsScanning(false)
          onScan(decodedText)
        })
      },
      () => {} // Ignore scan errors (no QR in frame)
    ).then(() => {
      setIsScanning(true)
    }).catch((err) => {
      onError?.(err.toString())
    })

    return () => {
      if (scannerRef.current?.isScanning) {
        scannerRef.current.stop().catch(() => {})
      }
    }
  }, [onScan, onError])

  return (
    <div className="flex flex-col items-center gap-4">
      <div
        id={containerRef.current}
        className="w-full max-w-sm rounded-2xl overflow-hidden"
      />
      {isScanning && (
        <p className="text-sm text-gray-400">Scanning for QR code...</p>
      )}
    </div>
  )
}
