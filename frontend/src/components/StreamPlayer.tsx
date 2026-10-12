import { useEffect, useRef } from "react"


type StreamPlayerProps = {
    wsUrl: string
    isPlaying: boolean
    onStatusChange: (status: 'connecting' | 'live' |'error') => void
}



const StreamPlayer = ({ wsUrl, isPlaying, onStatusChange }: StreamPlayerProps) => {

    const canvasRef = useRef<HTMLCanvasElement>(null)

    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas || !isPlaying) return
      
        let disposed = false
        let firstFrameDecoded = false
       
        
        
        onStatusChange('connecting')
      
        const player = new JSMpeg.Player(wsUrl, {
          canvas,
          audio: false,
          reconnectInterval: 0,
          disableWebAssembly: true,
          pauseWhenHidden: false,
          onVideoDecode: () => {
            if (disposed || firstFrameDecoded) return
      
            firstFrameDecoded = true
            clearTimeout(startupTimer)
            onStatusChange('live')
          },
        })
      
        const socket = player.source.socket
      
        function dispose() {
          if (disposed) return
          disposed = true
      
          clearTimeout(startupTimer)
          socket.removeEventListener('error', handleFailure)
          socket.removeEventListener('close', handleFailure)
          player.destroy()
        }
      
        function handleFailure() {
          if (disposed) return
          dispose()
          onStatusChange('error')
        }
      
        socket.addEventListener('error', handleFailure)
        socket.addEventListener('close', handleFailure)
      
        const startupTimer = setTimeout(handleFailure, 15_000)
      
        return dispose
      }, [wsUrl, isPlaying, onStatusChange])

    return <canvas ref={canvasRef} aria-label="Live video stream" />

}

export default StreamPlayer
