import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ReceivePage } from './pages/ReceivePage'
import { PayPage } from './pages/PayPage'
import { HistoryPage } from './pages/HistoryPage'
import { MerchantPOS } from './pages/MerchantPOS'

function Home() {
  return (
    <div className="flex flex-col items-center justify-center pt-20 gap-6">
      <h1 className="text-4xl font-bold">nostr-pay</h1>
      <p className="text-gray-400 text-center max-w-xs">
        Instant Lightning payments powered by Nostr
      </p>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Home />} />
          <Route path="/receive" element={<ReceivePage />} />
          <Route path="/pay" element={<PayPage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Route>
        <Route path="/merchant/pos" element={<MerchantPOS />} />
      </Routes>
    </BrowserRouter>
  )
}
