import { useState } from 'react'
import { cn } from '../../lib/cn'
import { Skeleton } from './Skeleton'

type FastImageProps = {
  uri: string
  alt: string
  width?: number | string
  height?: number | string
  radius?: number
  className?: string
  imgClassName?: string
}

export function FastImage({
  uri,
  alt,
  width = '100%',
  height = 160,
  radius = 12,
  className,
  imgClassName,
}: FastImageProps) {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  return (
    <div
      className={cn('relative overflow-hidden bg-slate-200', className)}
      style={{
        width: className ? undefined : width,
        height: className ? undefined : height,
        borderRadius: radius,
      }}
    >
      {!loaded && !failed ? (
        <Skeleton width="100%" height="100%" radius={radius} className="absolute inset-0 h-full" />
      ) : null}
      {failed ? (
        <div className="flex h-full items-center justify-center px-3 text-center text-sm text-slate-500">{alt}</div>
      ) : (
        <img
          alt={alt}
          src={uri}
          loading="lazy"
          decoding="async"
          onLoad={() => setLoaded(true)}
          onError={() => setFailed(true)}
          className={cn(
            'h-full w-full object-cover transition-opacity duration-300',
            imgClassName,
          )}
          style={{ opacity: loaded ? 1 : 0, position: loaded ? 'relative' : 'absolute', inset: 0 }}
        />
      )}
    </div>
  )
}
