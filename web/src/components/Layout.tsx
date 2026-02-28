import { useState } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../stores/auth'
import { LoginModal } from './LoginModal'

const navItems = [
  { path: '/receive', label: 'Receive', icon: '↓' },
  { path: '/pay', label: 'Pay', icon: '↑' },
  { path: '/history', label: 'History', icon: '☰' },
  { path: '/merchant/pos', label: 'POS', icon: '◻' },
]

export function Layout() {
  const location = useLocation()
  const { isLoggedIn, pubkey, logout } = useAuth()
  const [showLogin, setShowLogin] = useState(false)

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      <header className="px-4 py-3 border-b border-gray-800 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold">nostr-pay</Link>
        <div className="flex items-center gap-3">
          {isLoggedIn ? (
            <>
              <span className="text-xs text-gray-400 font-mono">
                {pubkey?.slice(0, 8)}...{pubkey?.slice(-4)}
              </span>
              <button
                onClick={logout}
                className="text-sm text-gray-400 hover:text-white border border-gray-700 px-3 py-1 rounded-lg"
              >
                Logout
              </button>
            </>
          ) : (
            <button
              onClick={() => setShowLogin(true)}
              className="text-sm bg-amber-500 hover:bg-amber-600 text-black font-bold px-4 py-1.5 rounded-lg"
            >
              Login
            </button>
          )}
        </div>
      </header>

      <main className="flex-1 p-4">
        <Outlet />
      </main>

      <nav className="border-t border-gray-800 px-4 py-2">
        <div className="flex justify-around">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex flex-col items-center py-2 px-3 rounded-lg text-sm ${
                location.pathname === item.path
                  ? 'text-amber-400'
                  : 'text-gray-500 hover:text-gray-300'
              }`}
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </Link>
          ))}
        </div>
      </nav>

      {showLogin && <LoginModal onClose={() => setShowLogin(false)} />}
    </div>
  )
}
