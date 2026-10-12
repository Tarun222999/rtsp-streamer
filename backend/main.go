package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coder/websocket"
)

var streamSlots = make(chan struct{}, 4)

var (
	streamMu     sync.Mutex
	streamWG     sync.WaitGroup
	shuttingDown bool
)

func main() {
	appCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	configureDemoStreams()

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/ws", handleWebSocket)

	port := envOrDefault("PORT", "8080")

	server := &http.Server{
		Addr:              ":" + port,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return appCtx
		},
	}

	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)
		<-appCtx.Done()

		streamMu.Lock()
		shuttingDown = true
		streamMu.Unlock()

		log.Println("Shutdown requested")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown failed: %v", err)
			server.Close()
		}

		streamsDone := make(chan struct{})
		go func() {
			streamWG.Wait()
			close(streamsDone)
		}()

		select {
		case <-streamsDone:
			log.Println("Stream cleanup complete")
		case <-shutdownCtx.Done():
			log.Println("Shutdown deadline reached")
		}
	}()

	log.Printf("Server listening on :%s", port)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Printf("HTTP server failed: %v", err)
	}

	stop()
	<-shutdownDone
}

var allowedDemoStreams = make(map[string]struct{})

func configureDemoStreams() {
	base := strings.TrimRight(
		envOrDefault("RTSP_BASE_URL", "rtsp://127.0.0.1:8554"),
		"/",
	)

	for camera := 1; camera <= 4; camera++ {
		sourceURL := fmt.Sprintf("%s/camera%d", base, camera)
		allowedDemoStreams[sourceURL] = struct{}{}
	}
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

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
