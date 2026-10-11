import StreamPlayer from './components/StreamPlayer'
import './App.css'
import { useState } from 'react'

function App() {
  const [isPlaying, setIsPlaying] = useState(true)

  return (
    <main>
      <h1>Camera 1</h1>
      <StreamPlayer
        wsUrl="ws://localhost:8080/ws"
        isPlaying={isPlaying}
      />
      {
        <button type='button' onClick={()=>setIsPlaying((isPlaying) => !isPlaying)}>
          {isPlaying ? 'Pause': 'Play'}
        </button>
      }
    </main>
  )
}

export default App