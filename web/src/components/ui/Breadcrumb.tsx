import { Link } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'
import { ProjectIcon } from '@/components/projects/ProjectIcon'
import type { Project } from '@/types/project'

export interface BreadcrumbItem {
  label: string
  href?: string  // undefined means current page (no link)
}

interface BreadcrumbProps {
  items: BreadcrumbItem[]
  project?: Project  // プロジェクトが渡された場合、先頭にアイコンを表示
}

export function Breadcrumb({ items, project }: BreadcrumbProps) {
  if (items.length === 0) return null

  return (
    <nav className="mb-4 flex items-center gap-2 text-sm text-gray-500">
      {project && <ProjectIcon project={project} className="h-5 w-5" />}
      {items.map((item, index) => (
        <span key={index} className="flex items-center gap-1">
          {index > 0 && <ChevronRight className="h-4 w-4 text-gray-400" />}
          {item.href ? (
            <Link to={item.href} className="hover:text-gray-700 truncate max-w-[200px]">
              {item.label}
            </Link>
          ) : (
            <span className="text-gray-900 font-medium truncate max-w-[200px]">
              {item.label}
            </span>
          )}
        </span>
      ))}
    </nav>
  )
}
