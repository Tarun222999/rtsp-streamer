import { useState } from 'react'
import StreamPlayer from './StreamPlayer'

type StreamCardProps = {
  title: string
  wsUrl: string
  onRemove: () => void
}

function StreamCard({ title, wsUrl, onRemove }: StreamCardProps) {
  const [isPlaying, setIsPlaying] = useState(true)
  const [status, setStatus] =
    useState<'connecting' | 'live' | 'error'>('connecting')
  const [retryAttempt, setRetryAttempt] = useState(0)

  return (
    <article>
      <h2>{title}</h2>

      <p role="status">
        {!isPlaying
          ? 'Paused'
          : status === 'live'
            ? 'Live'
            : status === 'error'
              ? 'Stream unavailable or disconnected. Try again.'
              : 'Connecting…'}
      </p>
      <StreamPlayer
        wsUrl={wsUrl}
        isPlaying={isPlaying}
        onStatusChange={setStatus}
        key={retryAttempt}
      />

      <button type="button" onClick={() => setIsPlaying(!isPlaying)}>
        {isPlaying ? 'Pause' : 'Play'}
      </button>

      <button type="button" onClick={onRemove}>
        Remove
      </button>

      {isPlaying && status === 'error' && (
        <button
          type="button"
          onClick={() => setRetryAttempt((attempt) => attempt + 1)}
        >
          Retry
        </button>
      )}
    </article>
  )
}

export default StreamCard
