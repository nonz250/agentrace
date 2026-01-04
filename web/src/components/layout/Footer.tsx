import { useQuery } from '@tanstack/react-query'
import { Github, Coffee } from 'lucide-react'
import { getVersion } from '@/api/version'

export function Footer() {
  const { data: versionInfo } = useQuery({
    queryKey: ['version'],
    queryFn: getVersion,
    staleTime: Infinity,
  })

  return (
    <footer className="border-t border-gray-200 bg-gray-50 py-4">
      <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-2 px-4 text-sm text-gray-500 sm:flex-row sm:gap-4">
        <div className="flex items-center gap-2">
          <a
            href="https://github.com/satetsu888/agentrace"
            target="_blank"
            rel="noopener noreferrer"
            className="text-gray-400 transition-colors hover:text-gray-600"
            aria-label="GitHub Repository"
          >
            <Github className="h-5 w-5" />
          </a>
          <span className="font-medium">agentrace</span>
          {versionInfo && (
            <span className="text-gray-400">{versionInfo.version}</span>
          )}
        </div>
        <a
          href="https://buymeacoffee.com/satetsu888"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1 text-gray-400 transition-colors hover:text-gray-600"
        >
          <Coffee className="h-4 w-4" />
          <span>Buy me a coffee</span>
        </a>
      </div>
    </footer>
  )
}
