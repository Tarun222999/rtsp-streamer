
import { type FormEvent, useState } from 'react'
import './App.css'
import StreamCard from './components/StreamCard';

function createStreamWebSocketUrl(sourceUrl: string) {
  const websocketUrl = new URL('/ws', import.meta.env.VITE_BACKEND_URL)
  websocketUrl.protocol =
    websocketUrl.protocol === 'https:' ? 'wss:' : 'ws:'
  websocketUrl.searchParams.set('url', sourceUrl)

  return websocketUrl.toString()
}

const demoStreams = [
  {
    id: 'camera-1',
    title: 'Camera 1',
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

const initialStreams = demoStreams.map(({ id, ...stream }) => ({
  ...stream,
  id: `viewer-${id}`,
}))

function App() {
  const [streams, setStreams] = useState(initialStreams)
  const [draftUrl, setDraftUrl] = useState('')
  const [selectedDemoUrl, setSelectedDemoUrl] = useState('')
  const [formError, setFormError] = useState<string | null>(null)

  function removeStream(id: string) {
    setStreams((currentStreams) =>
      currentStreams.filter((stream) => stream.id !== id),
    )
  }

  function addStream(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const sourceUrl = draftUrl.trim()
    if (!sourceUrl) {
      setFormError('Enter an RTSP URL or choose a demo stream.')
      return
    }

    const selectedDemo = demoStreams.find(
      (stream) => stream.sourceUrl === sourceUrl,
    )

    if (!selectedDemo) {
      setFormError('Only Camera 1 through Camera 4 are currently supported.')
      return
    }

    if (streams.some((stream) => stream.sourceUrl === sourceUrl)) {
      setFormError('That stream is already on screen.')
      return
    }

    setStreams((currentStreams) => [
      ...currentStreams,
      { ...selectedDemo, id: crypto.randomUUID() },
    ])
    setDraftUrl('')
    setSelectedDemoUrl('')
    setFormError(null)
  }

  return (
    <main className="app-shell">
      <h1>RTSP Stream Viewer</h1>

      <form className="stream-form" onSubmit={addStream}>
        <div className="stream-form__field">
          <label htmlFor="demo-stream">Available stream</label>
          <select
            id="demo-stream"
            value={selectedDemoUrl}
            onChange={(event) => {
              const sourceUrl = event.target.value
              setSelectedDemoUrl(sourceUrl)
              setDraftUrl(sourceUrl)
              setFormError(null)
            }}
          >
            <option value="">Choose a supported stream</option>
            {demoStreams.map((stream) => (
              <option
                key={stream.id}
                value={stream.sourceUrl}
                disabled={streams.some(
                  (activeStream) => activeStream.sourceUrl === stream.sourceUrl,
                )}
              >
                {stream.title}
              </option>
            ))}
          </select>
        </div>

        <div className="stream-form__field">
          <label htmlFor="stream-url">RTSP URL</label>
          <input
            id="stream-url"
            type="url"
            inputMode="url"
            placeholder="rtsp://127.0.0.1:8554/camera1"
            value={draftUrl}
            onChange={(event) => {
              setDraftUrl(event.target.value)
              setSelectedDemoUrl('')
              setFormError(null)
            }}
          />
        </div>

        <button type="submit">Add stream</button>
        <p className="stream-form__help">
          Choose a supported camera feed or enter its RTSP URL.
        </p>
        {formError && <p className="stream-form__error">{formError}</p>}
      </form>

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
