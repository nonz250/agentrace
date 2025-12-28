import { cn } from '@/lib/cn'

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  hover?: boolean
}

export function Card({ className, hover, ...props }: CardProps) {
  return (
    <div
      className={cn(
        'rounded-xl border border-gray-200 bg-white p-4 shadow-sm',
        hover && 'cursor-pointer transition-shadow hover:shadow-md',
        className
      )}
      {...props}
    />
  )
}
