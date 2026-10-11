
import { useState } from 'react';
import './App.css'
import StreamCard from './components/StreamCard';



function createStreamWebSocketUrl(sourceUrl: string) {
  const websocketUrl = new URL('/ws', import.meta.env.VITE_BACKEND_URL)
  websocketUrl.protocol =
    websocketUrl.protocol === 'https:' ? 'wss:' : 'ws:'
  websocketUrl.searchParams.set('url', sourceUrl)

  return websocketUrl.toString()
}


const initialStreams = [
  {
    id: 'viewer-a',
    title: 'Camera 1 — Viewer A',
    sourceUrl: 'rtsp://127.0.0.1:8554/camera1',
  },
  {
    id: 'viewer-b',
    title: 'Camera 2',
    sourceUrl: 'rtsp://127.0.0.1:8554/camera2',
  },
  {
    id: 'viewer-c',
    title: 'Camera 3',
    sourceUrl: 'rtsp://127.0.0.1:8554/camera3',
  },
  {
    id: 'viewer-d',
    title: 'Camera 4',
    sourceUrl: 'rtsp://127.0.0.1:8554/camera4',
  },
]

function App() {

  const [streams, setStreams] = useState(initialStreams)

  function removeStream(id: string) {
    setStreams((currentStreams) =>
      currentStreams.filter((stream) => stream.id !== id),
    )
  }

  return (
    <main>
      <h1>RTSP Stream Viewer</h1>

      <section className="stream-grid">
        {streams.map((stream) => (
          <StreamCard
            key={stream.id}
            title={stream.title}
            wsUrl={createStreamWebSocketUrl(stream.sourceUrl)}
            onRemove={() => removeStream(stream.id)}
          />
        ))}
      </section>
    </main>

  )
}

export default App