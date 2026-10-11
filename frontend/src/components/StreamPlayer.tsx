import { useEffect, useRef } from "react"


type StreamPlayerProps = {
    wsUrl: string
    isPlaying:boolean
}



const StreamPlayer = ({ wsUrl,isPlaying }: StreamPlayerProps) => {

    const canvasRef = useRef<HTMLCanvasElement>(null)

    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas || !isPlaying) return

        const player = new JSMpeg.Player(wsUrl, {
            canvas,
            audio: false,
            reconnectInterval: 0,
            disableWebAssembly: true,
            pauseWhenHidden: false,
          })

          
        return () => {
            player.destroy()
        }
    }, [wsUrl,isPlaying])
    return <canvas ref={canvasRef} aria-label="Live video stream" />

}

export default StreamPlayer
