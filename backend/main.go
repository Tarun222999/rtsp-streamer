package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
