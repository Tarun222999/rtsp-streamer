package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/coder/websocket"
)

var streamSlots = make(chan struct{}, 4)

var (
	streamMu     sync.Mutex
	streamWG     sync.WaitGroup
	shuttingDown bool
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	streamURL, err := validateStreamURL(r.URL.Query().Get("url"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	streamMu.Lock()

	if shuttingDown {
		streamMu.Unlock()
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
		return
	}

	streamWG.Add(1)
	streamMu.Unlock()
	defer streamWG.Done()

	select {
	case streamSlots <- struct{}{}:
		defer func() { <-streamSlots }()
	default:
		http.Error(
			w,
			"Stream capacity reached. Pause another player and retry.",
			http.StatusServiceUnavailable,
		)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{
			envOrDefault("FRONTEND_ORIGIN", "localhost:5173"),
		},
	})

	if err != nil {
		log.Printf("Websocket connect failed: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	//detex browser disconnects while we sending video
	ctx = conn.CloseRead(ctx)

	//convert H.264 into MPEG
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-rtsp_transport", "tcp",
		"-timeout", "10000000",
		"-i", streamURL,
		"-an",
		"-c:v", "mpeg1video",
		"-s", "320x240",
		"-r", "25",
		"-bf", "0",
		"-f", "mpegts",
		"pipe:1",
	)

	cmd.Stderr = os.Stderr

	output, err := cmd.StdoutPipe()

	if err != nil {
		log.Printf("Cannot open FFmpeg output: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Cannot start FFmpeg: %v", err)
		return
	}

	log.Println("Receiving FFmpeg started")

	defer func() {
		cancel()
		cmd.Wait()
		log.Println("Recieving FFmeg stopped")
	}()

	buffer := make([]byte, 32*1024)

	for {
		n, readErr := output.Read(buffer)

		if n > 0 {
			writeCtx, stopWrite := context.WithTimeout(ctx, 5*time.Second)

			err := conn.Write(
				writeCtx,
				websocket.MessageBinary,
				buffer[:n],
			)

			stopWrite()

			if err != nil {
				return
			}
		}

		if readErr != nil {
			if readErr != io.EOF && ctx.Err() == nil {
				log.Printf("FFmpeg output failed: %v", readErr)
			}
			return
		}
	}

}
