import { useQuery } from '@tanstack/react-query'
import { MemberList } from '@/components/members/MemberList'
import { Spinner } from '@/components/ui/Spinner'
import * as authApi from '@/api/auth'

export function MembersPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: authApi.getUsers,
  })

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
        Failed to load members: {error.message}
      </div>
    )
  }

  return (
    <div>
      <h1 className="mb-6 text-2xl font-semibold text-gray-900">Members</h1>
      <MemberList members={data?.users || []} />
    </div>
  )
}
