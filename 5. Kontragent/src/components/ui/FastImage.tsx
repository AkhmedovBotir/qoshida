import { useState } from 'react'
import { Skeleton } from './Skeleton'

type FastImageProps = {
  uri: string
  alt: string
  width?: number | string
  height?: number
  radius?: number
}

export function FastImage({
  uri,
  alt,
  width = '100%',
  height = 160,
  radius = 12,
}: FastImageProps) {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  return (
    <div
      className="relative overflow-hidden bg-slate-200"
      style={{ width, height, borderRadius: radius }}
    >
      {!loaded && !failed ? <Skeleton width="100%" height={height} radius={radius} /> : null}
      {failed ? (
        <div className="flex h-full items-center justify-center text-sm text-slate-500">{alt}</div>
      ) : (
        <img
          alt={alt}
          src={uri}
          loading="lazy"
          decoding="async"
          onLoad={() => setLoaded(true)}
          onError={() => setFailed(true)}
          className="h-full w-full object-cover transition-opacity duration-300"
          style={{ opacity: loaded ? 1 : 0, position: loaded ? 'relative' : 'absolute', inset: 0 }}
        />
      )}
    </div>
  )
}
