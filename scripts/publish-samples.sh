#!/bin/sh
set -eu

RTSP_BASE_URL="${RTSP_BASE_URL:-rtsp://mediamtx:8554}"
RTSP_BASE_URL="${RTSP_BASE_URL%/}"
pids=""

cleanup() {
    trap - EXIT INT TERM
    for pid in $pids; do
        kill "$pid" 2>/dev/null || true
    done
    wait || true
}

trap cleanup EXIT
trap 'exit 0' INT TERM

for camera in 1 2 3 4; do
    file="/media/camera${camera}.mp4"

    if [ ! -f "$file" ]; then
        echo "Missing sample: $file" >&2
        exit 1
    fi

    (
        while true; do
            ffmpeg \
                -hide_banner \
                -loglevel error \
                -nostdin \
                -re \
                -stream_loop -1 \
                -i "$file" \
                -map 0:v:0 \
                -an \
                -c:v copy \
                -f rtsp \
                -rtsp_transport tcp \
                "${RTSP_BASE_URL}/camera${camera}" || true

            sleep 2
        done
    ) &

    pids="$pids $!"
done

wait