
import { useState } from 'react';
import './App.css'
import StreamCard from './components/StreamCard';



const streamWebSocketUrl = new URL('/ws', import.meta.env.VITE_BACKEND_URL);
streamWebSocketUrl.protocol = streamWebSocketUrl.protocol === 'https:' ? 'wss:' : 'ws:'

const initialStreams  = [
  { id: 'viewer-a', title: 'Camera 1 — Viewer A' },
  { id: 'viewer-b', title: 'Camera 1 — Viewer B' },
]


function App() {

  const [streams,setStreams]=useState(initialStreams)

  function removeStream(id: string) {
    setStreams((currentStreams) =>
      currentStreams.filter((stream) => stream.id !== id),
    )
  }

  return (
    <main>
      <h1>RTSP Stream Viewer</h1>

      <section  className="stream-grid">
        {streams.map((stream) => (
          <StreamCard
            key={stream.id}
            title={stream.title}
            wsUrl={streamWebSocketUrl.toString()}
            onRemove={()=>removeStream(stream.id)}
          />
        ))}
      </section>
    </main>

  )
}

export default App