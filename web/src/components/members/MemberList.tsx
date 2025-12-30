import { User as UserIcon, Mail, Calendar } from 'lucide-react'
import { format } from 'date-fns'
import { Card } from '@/components/ui/Card'
import type { User } from '@/types/auth'

interface MemberListProps {
  members: User[]
}

export function MemberList({ members }: MemberListProps) {
  if (members.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-gray-300 bg-white p-6 text-center">
        <p className="text-gray-500">No members registered.</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {members.map((member) => (
        <Card key={member.id}>
          <div className="flex items-start gap-3">
            <UserIcon className="mt-0.5 h-5 w-5 flex-shrink-0 text-gray-400" />
            <div>
              <p className="font-medium text-gray-900">
                {member.display_name || member.email}
              </p>
              <div className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm text-gray-500">
                <span className="flex items-center gap-1">
                  <Mail className="h-3.5 w-3.5" />
                  {member.email}
                </span>
                <span className="flex items-center gap-1">
                  <Calendar className="h-3.5 w-3.5" />
                  Joined {format(new Date(member.created_at), 'yyyy/MM/dd')}
                </span>
              </div>
            </div>
          </div>
        </Card>
      ))}
    </div>
  )
}
