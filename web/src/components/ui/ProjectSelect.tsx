import { useQuery } from '@tanstack/react-query'
import { Select } from './Select'
import { getProjects } from '@/api/projects'
import { getProjectDisplayName } from '@/lib/project-utils'

interface ProjectSelectProps {
  value: string
  onChange: (projectId: string) => void
  disabled?: boolean
  className?: string
  showOptional?: boolean
}

export function ProjectSelect({
  value,
  onChange,
  disabled = false,
  className,
  showOptional = true,
}: ProjectSelectProps) {
  const { data: projectsData, isLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => getProjects(),
  })

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange(e.target.value)
  }

  return (
    <Select
      value={value}
      onChange={handleChange}
      disabled={disabled || isLoading}
      className={className}
    >
      {showOptional && <option value="">Select a project (optional)</option>}
      {projectsData?.projects.map((project) => (
        <option key={project.id} value={project.id}>
          {getProjectDisplayName(project) || 'Default Project'}
        </option>
      ))}
    </Select>
  )
}
