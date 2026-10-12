declare namespace JSMpeg {
  interface PlayerOptions {
    canvas: HTMLCanvasElement
    audio?: boolean
    reconnectInterval?: number
    disableWebAssembly?: boolean
    pauseWhenHidden?: boolean
    onVideoDecode?: () => void
  }

  class Player {
    constructor(url: string, options: PlayerOptions)
    destroy(): void
    readonly source: {
      readonly socket: WebSocket
    }
  }
}