package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/coder/websocket"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

var allowedDemoStreams = map[string]struct{}{
	"rtsp://127.0.0.1:8554/camera1": {},
	"rtsp://127.0.0.1:8554/camera2": {},
	"rtsp://127.0.0.1:8554/camera3": {},
	"rtsp://127.0.0.1:8554/camera4": {},
}

func validateStreamURL(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf("missing url parameter")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid stream URL")
	}

	if parsed.Scheme != "rtsp" || parsed.Host == "" {
		return "", fmt.Errorf("stream URL must be a complete rtsp URL")
	}

	if _, allowed := allowedDemoStreams[raw]; !allowed {
		return "", fmt.Errorf("this stream is not allowed in the local demo")
	}

	return raw, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	streamURL, err := validateStreamURL(r.URL.Query().Get("url"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:5173"},
	})

	if err != nil {
		log.Printf("Websocket connect failed: %v", err)
		return
	}
	defer conn.CloseNow()

	//limit each prototype connection limit to one minute
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	//detex browser disconnects while we sending video
	ctx = conn.CloseRead(ctx)

	//convert H.264 into MPEG
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-rtsp_transport", "tcp",
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
