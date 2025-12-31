import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { PlanList } from '@/components/plans/PlanList'
import { Spinner } from '@/components/ui/Spinner'
import { Button } from '@/components/ui/Button'
import * as plansApi from '@/api/plan-documents'

const PAGE_SIZE = 20

export function PlansPage() {
  const [page, setPage] = useState(1)
  const offset = (page - 1) * PAGE_SIZE

  const { data, isLoading, error } = useQuery({
    queryKey: ['plans', 'list', page],
    queryFn: () => plansApi.getPlans({ limit: PAGE_SIZE, offset }),
  })

  const plans = data?.plans || []
  const hasMore = plans.length === PAGE_SIZE

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
        Failed to load plans: {error.message}
      </div>
    )
  }

  return (
    <div>
      <h1 className="mb-6 text-2xl font-semibold text-gray-900">Plans</h1>
      <PlanList plans={plans} />

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
