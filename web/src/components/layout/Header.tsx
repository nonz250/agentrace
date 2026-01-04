import { Link, useLocation, matchPath } from 'react-router-dom'
import { Settings, LogOut, ChevronDown, Users, MessageSquare, FileText } from 'lucide-react'
import { useState, useRef, useEffect, useMemo } from 'react'
import { useAuth } from '@/hooks/useAuth'

function useCurrentProjectId(): string | undefined {
  const location = useLocation()
  // /projects/:projectId/* にマッチするか確認
  const match = matchPath('/projects/:projectId/*', location.pathname)
  return match?.params.projectId
}

export function Header() {
  const { user, logout, isLoggingOut } = useAuth()
  const location = useLocation()
  const currentProjectId = useCurrentProjectId()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  // プロジェクト配下にいるときのみSessions/Plansリンクを表示
  const navLinks = useMemo(() => {
    if (currentProjectId) {
      return [
        { to: `/projects/${currentProjectId}/sessions`, label: 'Sessions', icon: MessageSquare },
        { to: `/projects/${currentProjectId}/plans`, label: 'Plans', icon: FileText },
      ]
    }
    // TOPページではリンクを表示しない
    return []
  }, [currentProjectId])

  const isActive = (to: string) => {
    return location.pathname.startsWith(to)
  }

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
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2 text-lg font-semibold text-gray-900">
            <span className="text-primary-600">&#9671;</span>
            Agentrace
          </Link>
          <nav className="flex items-center gap-6">
            {navLinks.map((link) => {
              const Icon = link.icon
              return (
                <Link
                  key={link.to}
                  to={link.to}
                  className={`flex items-center gap-1 text-sm font-medium transition-colors ${
                    isActive(link.to)
                      ? 'text-primary-600'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  <Icon className="h-4 w-4" />
                  {link.label}
                </Link>
              )
            })}
          </nav>
        </div>

        {user ? (
          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setMenuOpen(!menuOpen)}
              className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
            >
              {user.display_name || user.email}
              <ChevronDown className="h-4 w-4" />
            </button>

            {menuOpen && (
              <div className="absolute right-0 mt-2 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg">
                <Link
                  to="/members"
                  onClick={() => setMenuOpen(false)}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                >
                  <Users className="h-4 w-4" />
                  Members
                </Link>
                <Link
                  to="/settings"
                  onClick={() => setMenuOpen(false)}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                >
                  <Settings className="h-4 w-4" />
                  Settings
                </Link>
                <div className="my-1 border-t border-gray-200" />
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
        ) : (
          <div className="flex items-center gap-3">
            <Link
              to="/login"
              className="text-sm text-gray-600 hover:text-gray-900"
            >
              Login
            </Link>
            <Link
              to="/register"
              className="rounded-lg bg-primary-600 px-3 py-1.5 text-sm text-white hover:bg-primary-700"
            >
              Register
            </Link>
          </div>
        )}
      </div>
    </header>
  )
}
