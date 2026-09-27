import { cn } from '../../lib/cn'

type SkeletonProps = {
  width?: number | string
  height?: number | string
  radius?: number
  circle?: boolean
  className?: string
}

export function Skeleton({
  width = '100%',
  height = 16,
  radius = 10,
  circle = false,
  className,
}: SkeletonProps) {
  return (
    <div
      className={cn('skeleton', className)}
      style={{
        width,
        height: circle ? width : height,
        borderRadius: circle ? 999 : radius,
      }}
    />
  )
}

export function SkeletonCard() {
  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-3.5">
      <Skeleton height={140} radius={12} />
      <Skeleton width="70%" height={14} />
      <Skeleton width="40%" height={12} />
    </div>
  )
}
