import { Link } from 'react-router-dom'
import { Settings, LogOut, ChevronDown } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'
import { useAuth } from '@/hooks/useAuth'

export function Header() {
  const { user, logout, isLoggingOut } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <header className="sticky top-0 z-10 border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
        <Link to="/" className="flex items-center gap-2 text-lg font-semibold text-gray-900">
          <span className="text-primary-600">&#9671;</span>
          Agentrace
        </Link>

        <div className="flex items-center gap-4">
          <Link
            to="/settings"
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
          >
            <Settings className="h-4 w-4" />
            Settings
          </Link>

          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setMenuOpen(!menuOpen)}
              className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
            >
              {user?.name}
              <ChevronDown className="h-4 w-4" />
            </button>

            {menuOpen && (
              <div className="absolute right-0 mt-2 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg">
                <button
                  onClick={() => {
                    logout()
                    setMenuOpen(false)
                  }}
                  disabled={isLoggingOut}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                >
                  <LogOut className="h-4 w-4" />
                  {isLoggingOut ? 'Logging out...' : 'Logout'}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
