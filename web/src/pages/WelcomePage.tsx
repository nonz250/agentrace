import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'

export function WelcomePage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md text-center">
        <h1 className="mb-2 flex items-center justify-center gap-2 text-3xl font-bold text-gray-900">
          <span className="text-primary-600">&#9671;</span>
          Agentrace
        </h1>
        <p className="mb-8 text-gray-600">
          Track and review Claude Code sessions with your team.
        </p>

        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Link to="/register">
            <Button size="lg" className="w-full sm:w-auto">
              Register
            </Button>
          </Link>
          <Link to="/login">
            <Button variant="secondary" size="lg" className="w-full sm:w-auto">
              Login with API Key
            </Button>
          </Link>
        </div>
      </div>
    </div>
  )
}
