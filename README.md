# RTSP Stream Viewer

A React/TypeScript and Go application for watching live RTSP video in a browser. Add a configured stream URL, watch multiple sources in a responsive grid, and independently pause, resume, retry or remove each player.

**Live demo:** Not deployed yet. This README will be updated with the public URL after deployment verification.

## Features

- Four distinct, looping test sources hosted by MediaMTX.
- Stream selection and RTSP URL input; players start only after adding a stream.
- Responsive grid with independent Play/Pause, Remove and Retry controls.
- Connecting, Live, Paused and Error status indicators.
- Go WebSocket delivery of FFmpeg-generated MPEG-TS video, decoded with JSMpeg.
- Source allowlist, four concurrent receiver slots and bounded socket operations.
- FFmpeg cleanup on disconnect and coordinated server shutdown.
- Docker setup for the backend, media server and test publishers.

## Architecture

```text
Four sample MP4s
    → four publishing FFmpeg processes (loop, stream copy)
    → MediaMTX /camera1 ... /camera4 (RTSP)
    → Go backend + one receiving FFmpeg per active player
    → binary MPEG-TS over WebSocket
    → React + JSMpeg canvas player
```

Publishing is independent of viewers. MediaMTX serves the live sources; the backend reads a selected source and transcodes H.264 into MPEG-1 video for the browser decoder. The browser connects to Go, not directly to MediaMTX.

## Prerequisites

- Docker Desktop with the engine running and Docker Compose available.
- Node.js compatible with the Vite version in `frontend/package.json`, and npm.
- Git.

Go and FFmpeg are installed in the Docker build/runtime images. Host Go 1.27 and FFmpeg are needed only if running the backend outside Docker.

## Local setup

Clone the repository:

```powershell
git clone https://github.com/Tarun222999/rtsp-streamer.git
Set-Location rtsp-streamer
```

Build and start the backend and test-source services:

```powershell
docker compose up --build -d
docker compose ps
```

The publisher container includes `dev/media/camera1.mp4` through `camera4.mp4`. It retries if MediaMTX is not ready. Stop any host Go server on port 8080 or old test publishers before using this setup.

Create `frontend/.env.development.local` with:

```dotenv
VITE_BACKEND_URL=http://localhost:8080
VITE_RTSP_BASE_URL=rtsp://mediamtx:8554
```

Start the frontend in a separate terminal:

```powershell
Set-Location frontend
npm ci
npm run dev -- --port 5173 --strictPort
```

Open [http://localhost:5173](http://localhost:5173). Use this exact hostname and port to match the configured WebSocket origin. Restart Vite after changing its environment variables.

The health endpoint is [http://localhost:8080/healthz](http://localhost:8080/healthz).

### Add a stream

Select a camera in the dropdown and click **Add stream**, or enter one of these URLs:

```text
rtsp://mediamtx:8554/camera1
rtsp://mediamtx:8554/camera2
rtsp://mediamtx:8554/camera3
rtsp://mediamtx:8554/camera4
```

These URLs are resolved by the backend inside Docker. A browser does not need to resolve the `mediamtx` hostname.

- **Pause:** destroys that player's decoder/socket and stops its receiving FFmpeg.
- **Play:** opens a fresh connection at the current live position.
- **Remove:** removes the card and releases its receiver.
- **Retry:** recreates a failed player.

Pausing skips footage while paused. Other players and the test publishers keep running.

### Logs and shutdown

```powershell
docker compose logs --tail 50 backend publishers mediamtx
docker compose down
```

Compose allows fifteen seconds for container termination. The backend gives coordinated HTTP and stream cleanup a ten-second budget. Tini forwards termination signals; the publisher image uses process-group forwarding for its child processes.

## Configuration

| Variable | Component | Purpose |
| --- | --- | --- |
| `PORT` | Backend | HTTP listening port; default `8080` |
| `FRONTEND_ORIGIN` | Backend | Allowed frontend hostname and optional port, without scheme; default `localhost:5173` |
| `RTSP_BASE_URL` | Backend | Base address for the four allowed camera paths; host default `rtsp://127.0.0.1:8554` |
| `RTSP_BASE_URL` | Publishers | MediaMTX destination; default `rtsp://mediamtx:8554` |
| `VITE_BACKEND_URL` | Frontend | Backend HTTP/HTTPS base URL; converted to WS/WSS for streaming |
| `VITE_RTSP_BASE_URL` | Frontend | Base address submitted to the backend; must match the backend's allowed sources |

Vite variables are embedded during development startup or production build and are public configuration. Do not put secrets in them.

## WebSocket contract

Each active player opens its own connection:

```text
/ws?url=<URL-encoded RTSP source>
```

The backend validates the source before upgrading. Missing, invalid or disallowed URLs return HTTP 400; unavailable receiver capacity returns HTTP 503. Successful connections deliver binary MPEG-TS chunks, with no client routing messages required.

Browser WebSockets do not expose failed-handshake response bodies, so the frontend displays a general unavailable/disconnected message. The frontend allows fifteen seconds for the first decoded frame; FFmpeg's RTSP TCP socket timeout is ten seconds and WebSocket writes have a five-second timeout.

## Checks

Frontend, from `frontend/`:

```powershell
npm run lint
npm run build
```

Backend, from `backend/`:

```powershell
go build ./...
go vet ./...
```

For manual verification, add all four sources, watch beyond one minute, pause/remove one player while the others continue, retry after a source interruption, and inspect cleanup logs when stopping the backend. These are verification instructions, not a claim that an automated test suite exists.

## Deployment preparation

The intended deployment is a static frontend on Vercel and container services on Railway or a comparable platform. Production platform configuration and public verification are still pending.

1. Deploy MediaMTX and the publisher image on a private network. Build publishers using the repository root as context and `dev/Dockerfile.publishers` as the Dockerfile.
2. Deploy the backend using `backend/` as build context and its Dockerfile. Configure the platform's HTTP port, the exact frontend hostname and the private MediaMTX RTSP address.
3. Configure publishers to use that same private media-server address.
4. Build the frontend with `VITE_BACKEND_URL` set to the backend's HTTPS URL and `VITE_RTSP_BASE_URL` matching the backend configuration. Use `frontend/` as project root, `npm run build` and `dist` as output.
5. Verify secure WebSockets, four sources, controls, errors, restart cleanup and hosting usage limits. Add the verified live demo URL above.

The local Compose hostname `mediamtx` is not automatically the deployed platform's private hostname. Local loopback port mappings are for development; production ingress must be configured on the hosting platform. Keep RTSP private and expose only the backend's HTTP/WebSocket service.

## Limitations and tradeoffs

- Only the four configured camera URLs are supported. Arbitrary external RTSP URLs are intentionally rejected in this demo; supporting trusted additional sources requires extending the policy and UI.
- Video only, at 320×240 and 25 fps. No audio, recording, seeking or replay.
- Four active receiver connections across all visitors per backend process, not four per visitor.
- One transcoder per active player. Shared workers for the same source would improve scalability but are outside this version.
- MPEG-1/JSMpeg provides a simple WebSocket playback pipeline at the cost of compression efficiency compared with newer codecs.
- No account authentication. The media setup and source policy are intended for an assignment demo.

## Troubleshooting

- **Connecting or Error:** check all services and publisher logs, and ensure frontend/backend RTSP base URLs match.
- **WebSocket rejected:** use `localhost:5173` locally or configure the deployed frontend hostname in `FRONTEND_ORIGIN`.
- **Capacity reached:** pause or remove active players in other tabs; the four-slot limit is shared.
- **Address already in use:** stop the host backend or conflicting containers before Compose startup.
- **Environment change has no effect:** restart Vite; production frontend environment changes require rebuilding.
- **Publisher script fails on Linux:** preserve LF line endings and UTF-8 without BOM in `scripts/publish-samples.sh`.

## Source layout

```text
backend/                 Go server, configuration, stream lifecycle, Dockerfile
frontend/                React/TypeScript app and JSMpeg assets
dev/media/               Four sample MP4s
dev/Dockerfile.publishers Publisher runtime image
scripts/                 Linux and host test publisher scripts
compose.yaml             Local container orchestration
project.md               Project plan
```

JSMpeg's license is included alongside its frontend and prototype assets as `jsmpeg.LICENSE`.
