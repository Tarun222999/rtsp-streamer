declare namespace JSMpeg {
  interface PlayerOptions {
    canvas: HTMLCanvasElement
    audio?: boolean
    reconnectInterval?: number
    disableWebAssembly?: boolean
    pauseWhenHidden?: boolean
  }

  class Player {
    constructor(url: string, options: PlayerOptions)
    destroy(): void
  }
}