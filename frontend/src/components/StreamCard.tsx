import { useState } from 'react'
import StreamPlayer from './StreamPlayer'

type StreamCardProps = {
  title:  string
  wsUrl: string
  onRemove:()=>void
}

function StreamCard({ title, wsUrl,onRemove }: StreamCardProps) {
  const [isPlaying, setIsPlaying] = useState(true)

  return (
    <article>
      <h2>{title}</h2>

      <StreamPlayer wsUrl={wsUrl} isPlaying={isPlaying} />

      <button type="button" onClick={() => setIsPlaying(!isPlaying)}>
        {isPlaying ? 'Pause' : 'Play'}
      </button>

      <button type="button" onClick={onRemove}>
        Remove
      </button>
    </article>
  )
}

export default StreamCard
