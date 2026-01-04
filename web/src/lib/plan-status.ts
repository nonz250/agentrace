import type { PlanDocumentStatus } from '@/types/plan-document'

export const statusConfig: Record<
  PlanDocumentStatus,
  { label: string; className: string }
> = {
  scratch: {
    label: 'Scratch',
    className: 'bg-orange-100 text-orange-700',
  },
  draft: {
    label: 'Draft',
    className: 'bg-gray-100 text-gray-600',
  },
  planning: {
    label: 'Planning',
    className: 'bg-blue-100 text-blue-700',
  },
  pending: {
    label: 'Pending',
    className: 'bg-yellow-100 text-yellow-700',
  },
  ready: {
    label: 'Ready',
    className: 'bg-cyan-100 text-cyan-700',
  },
  implementation: {
    label: 'Implementation',
    className: 'bg-purple-100 text-purple-700',
  },
  complete: {
    label: 'Complete',
    className: 'bg-green-100 text-green-700',
  },
}

// フィルターボタン用のborderクラスを追加
const statusBorderColors: Record<PlanDocumentStatus, string> = {
  scratch: 'border-orange-300',
  draft: 'border-gray-400',
  planning: 'border-blue-300',
  pending: 'border-yellow-300',
  ready: 'border-cyan-300',
  implementation: 'border-purple-300',
  complete: 'border-green-300',
}

const UNSELECTED_CLASS = 'bg-gray-50 text-gray-400 border-gray-200'

export function getFilterButtonClass(status: PlanDocumentStatus, isSelected: boolean): string {
  if (isSelected) {
    return `${statusConfig[status].className} ${statusBorderColors[status]}`
  }
  return UNSELECTED_CLASS
}
