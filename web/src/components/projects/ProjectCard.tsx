import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { MoreVertical, Trash2 } from 'lucide-react'
import type { Project } from '@/types/project'
import { parseRepoName, isDefaultProject, getRepoUrl } from '@/lib/project-utils'
import { ProjectIcon } from './ProjectIcon'
import { deleteProject } from '@/api/projects'
import { useAuth } from '@/hooks/useAuth'

interface ProjectCardProps {
  project: Project & { created_at: string }
  onDelete?: (id: string) => void
}

export function ProjectCard({ project, onDelete }: ProjectCardProps) {
  const { user } = useAuth()
  const repoName = parseRepoName(project)
  const repoUrl = getRepoUrl(project)
  const hasProject = !isDefaultProject(project)
  const [showMenu, setShowMenu] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const menuRef = useRef<HTMLDivElement>(null)

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setShowMenu(false)
      }
    }
    if (showMenu) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [showMenu])

  const handleMenuClick = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setShowMenu(!showMenu)
    setError(null)
  }

  const handleDelete = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()

    if (!confirm('Are you sure you want to delete this project?')) {
      setShowMenu(false)
      return
    }

    setIsDeleting(true)
    setError(null)
    try {
      await deleteProject(project.id)
      onDelete?.(project.id)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete project'
      setError(message)
    } finally {
      setIsDeleting(false)
      setShowMenu(false)
    }
  }

  return (
    <Link
      to={`/projects/${project.id}`}
      className="block rounded-xl border border-gray-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md relative"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 text-gray-900">
            <ProjectIcon project={project} className="h-5 w-5 flex-shrink-0" />
            {hasProject ? (
              <span className="font-medium truncate">{repoName}</span>
            ) : (
              <span className="font-medium text-gray-500">(no project)</span>
            )}
          </div>
          {hasProject && repoUrl && (
            <p className="mt-1 text-sm text-gray-500 truncate">
              {project.canonical_git_repository}
            </p>
          )}
          {error && (
            <p className="mt-1 text-sm text-red-500">{error}</p>
          )}
        </div>
        {user && hasProject && (
          <div className="relative ml-2" ref={menuRef}>
            <button
              onClick={handleMenuClick}
              className="p-1 text-gray-400 hover:text-gray-600 rounded hover:bg-gray-100"
              aria-label="Project options"
            >
              <MoreVertical className="h-4 w-4" />
            </button>
            {showMenu && (
              <div className="absolute right-0 top-full mt-1 w-32 rounded-md border border-gray-200 bg-white shadow-lg z-10">
                <button
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="flex w-full items-center gap-2 px-3 py-2 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50"
                >
                  <Trash2 className="h-4 w-4" />
                  {isDeleting ? 'Deleting...' : 'Delete'}
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </Link>
  )
}
