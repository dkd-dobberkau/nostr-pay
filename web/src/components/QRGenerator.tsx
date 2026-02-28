import { QRCodeSVG } from 'qrcode.react'

interface QRGeneratorProps {
  value: string
  size?: number
  label?: string
}

export function QRGenerator({ value, size = 256, label }: QRGeneratorProps) {
  return (
    <div className="flex flex-col items-center gap-4">
      <div className="bg-white p-4 rounded-2xl">
        <QRCodeSVG
          value={value}
          size={size}
          level="M"
          bgColor="#ffffff"
          fgColor="#000000"
        />
      </div>
      {label && (
        <p className="text-sm text-gray-400 text-center">{label}</p>
      )}
    </div>
  )
}
