import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight, Terminal, Copy, Check } from 'lucide-react'
import { ProjectList } from '@/components/projects/ProjectList'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import * as projectsApi from '@/api/projects'

const PAGE_SIZE = 20

function SetupGuide() {
  const [copied, setCopied] = useState(false)
  // Use VITE_API_URL if set (for dev), otherwise use current origin (for prod)
  const serverUrl = import.meta.env.VITE_API_URL || window.location.origin
  const command = `npx agentrace init --url ${serverUrl}`

  const handleCopy = async () => {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="mx-auto max-w-xl rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
      <div className="mb-6 text-center">
        <div className="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-full bg-primary-100">
          <Terminal className="h-8 w-8 text-primary-600" />
        </div>
        <h1 className="text-xl font-semibold text-gray-900">Welcome to Agentrace</h1>
        <p className="mt-2 text-gray-600">
          Get started by connecting Claude Code to this server.
        </p>
      </div>

      <div className="space-y-4">
        <div>
          <p className="mb-2 text-sm font-medium text-gray-700">
            Run this command in your terminal:
          </p>
          <div className="flex items-center gap-2 rounded-lg bg-gray-900 p-3">
            <code className="flex-1 font-mono text-sm text-gray-100">
              {command}
            </code>
            <button
              onClick={handleCopy}
              className="rounded p-1.5 text-gray-400 hover:bg-gray-800 hover:text-gray-200"
              title="Copy to clipboard"
            >
              {copied ? (
                <Check className="h-4 w-4 text-green-400" />
              ) : (
                <Copy className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        <p className="text-center text-sm text-gray-500">
          After setup, your Claude Code sessions will appear here.
        </p>
      </div>
    </div>
  )
}

export function ProjectsPage() {
  const [page, setPage] = useState(1)
  const offset = (page - 1) * PAGE_SIZE

  const { data, isLoading, error } = useQuery({
    queryKey: ['projects', 'list', page],
    queryFn: () => projectsApi.getProjects({ limit: PAGE_SIZE, offset }),
  })

  const projects = data?.projects || []
  const hasMore = projects.length === PAGE_SIZE

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
        Failed to load projects: {error.message}
      </div>
    )
  }

  // Check if there are real projects (not just the default empty project)
  const hasRealProjects = projects.some(
    (p) => p.id !== '00000000-0000-0000-0000-000000000000' && p.canonical_git_repository
  )

  // Show setup guide when there are no real projects
  if (!hasRealProjects) {
    return (
      <div className="py-12">
        <SetupGuide />
      </div>
    )
  }

  return (
    <div>
      <h1 className="mb-6 text-2xl font-semibold text-gray-900">Projects</h1>
      <ProjectList projects={projects} />

      {(page > 1 || hasMore) && (
        <div className="mt-6 flex items-center justify-between">
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <span className="text-sm text-gray-500">Page {page}</span>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setPage((p) => p + 1)}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  )
}
