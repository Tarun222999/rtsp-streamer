package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

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

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
