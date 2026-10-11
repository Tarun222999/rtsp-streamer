# RTSP Stream Viewer

## What this project does

Build a web application where users enter RTSP stream URLs and watch live video in their browser. Users can add multiple streams, view them in a responsive grid, and play or pause each player independently.

This project implements the **Skylark Labs Full Stack Engineer Coding Test**, using React and TypeScript for the frontend and Go for the backend.

## Required deliverables

- A GitHub repository containing the complete source code.
- A deployed frontend and backend with a working live demo.
- A README with local setup, deployment, and test-stream instructions.
- Any credentials needed to access the demo, if applicable.

The assignment specifies submission within seven days of receiving the test. The receipt date and submission deadline still need to be recorded.

## How streaming works

Browsers generally cannot play RTSP directly. The backend uses FFmpeg to read an RTSP source and convert its video into a format the frontend can decode. Go forwards that output over WebSockets.

### 1. Create a test source

```text
Local MP4 file → publishing FFmpeg → MediaMTX → RTSP URL
```

- Keep four small, distinct sample MP4 files in the test-stream setup, separate from application code. Different visuals make it easy to verify that each player receives the correct source.
- A publishing FFmpeg process loops that file and sends it to MediaMTX in real time.
- MediaMTX serves the incoming video at a configured path, such as `/camera1`.
- MediaMTX normally forwards and buffers the stream; recording it to disk is unnecessary.
- The test source runs while its publishing process is running, even without viewers.
- Use a suitably pre-encoded MP4 and stream copying where possible to reduce CPU usage.

### 2. Watch a source

```text
User enters RTSP URL
         ↓
React player connects to Go over WebSocket
         ↓
Go starts a receiving FFmpeg process
         ↓
FFmpeg reads RTSP from MediaMTX or a camera
         ↓
Go forwards browser-decodable video over WebSocket
         ↓
React player displays live video
```

The browser connects to the Go backend. It does not connect directly to MediaMTX. Entering a URL requests an existing stream; it does not start publishing the sample MP4.

A real RTSP camera can replace the test-source setup because the camera already exposes its own RTSP stream. The backend must be able to reach that camera's network.

## Planned stack

| Component | Technology | Purpose |
| --- | --- | --- |
| Frontend | React, TypeScript, Vite | URL input, stream grid, player controls and status |
| Backend | Go | Validate requests, manage FFmpeg processes, serve WebSockets |
| Video processing | FFmpeg | Read RTSP and produce browser-decodable video |
| Test media server | MediaMTX | Host our test RTSP streams |
| Local environment | Docker Desktop and Docker Compose | Run backend, MediaMTX and test publishers consistently |

### Initial playback approach

Start with MPEG-1 video in an MPEG-TS container over WebSockets, decoded by a JSMpeg player into a canvas. React owns the player lifecycle and controls. Prove this approach with one stream before building the complete interface.

This is an initial implementation choice, not an assignment requirement. Confirm decoder compatibility, browser performance and acceptable image quality in the first prototype. Audio is outside the initial scope because the assignment explicitly requires video streaming.

## Frontend behavior

- Input field for an RTSP URL and an Add Stream button.
- One player card for every added stream.
- Responsive grid: one column on small screens and multiple columns when space permits.
- Independent Play/Pause and Remove controls per card.
- Clear states: connecting, live, paused, error and disconnected.
- Useful errors for invalid URLs, unreachable sources, authentication failures and interrupted streams.
- Demo stream URLs available in the interface so reviewers can try the app easily.
- Hide embedded credentials from displayed URLs and error messages.

### Play and pause semantics

- **Play:** open a WebSocket connection and begin receiving live video.
- **Pause:** disconnect that player and preserve its last frame where practical.
- **Resume:** reconnect to the current live position; footage during the pause is skipped.
- **Remove:** disconnect the player and remove its card.

Pausing one player does not stop the test publisher or affect another player. Seeking, rewind and recording are outside the initial scope.

## Backend behavior

- Accept an RTSP URL and validate its scheme, host and permitted destination.
- Start FFmpeg with structured process arguments rather than a shell command.
- Forward binary video chunks through a WebSocket connection.
- Use one receiving FFmpeg process per active player connection in the initial version.
- Stop the associated FFmpeg process when that connection closes or fails.
- Bound concurrent streams, startup time and buffering so disconnected or slow clients do not exhaust resources.
- Report useful stream errors without exposing credentials or internal details.
- Provide a health endpoint and shut down child processes cleanly.
- Restrict allowed WebSocket origins in deployment.

Rooms are unnecessary for the initial version: each connection identifies the requested stream. Sharing one FFmpeg process among viewers of the same source is a possible later improvement.

### RTSP access policy

The public demo should allow only configured test sources to prevent arbitrary user URLs from accessing private infrastructure. Local development can explicitly allow the MediaMTX service or trusted camera destinations. Validate resolved destinations and account for hostname changes before allowing network access.

## Docker setup

Plan a Compose environment with:

1. **MediaMTX:** exposes RTSP sources inside the Docker network.
2. **Test publishers:** four FFmpeg processes, each looping its own sample video into a distinct path: `camera1`, `camera2`, `camera3` and `camera4`. All four use the same MediaMTX instance.
3. **Go backend:** includes FFmpeg and connects to MediaMTX through its Docker service name.

Run the React development server locally initially; containerizing it is optional.

URLs depend on where the receiving process runs:

- Backend inside Compose: `rtsp://mediamtx:8554/camera1`.
- Backend running directly on the host, with RTSP port published: `rtsp://localhost:8554/camera1`.

The backend consumes the submitted URL, so demo URLs must be reachable from the backend's environment. `localhost` on a deployed backend refers to that backend's own container or machine.

These components are separate processes and can share one machine. They do not require separate physical servers.

## Deployment plan

- Host the React frontend on Vercel or a comparable static host.
- Deploy the Go backend, MediaMTX and test publishers on Railway or a comparable container platform.
- Keep MediaMTX and publishers on private networking when hosted alongside the backend.
- Expose the backend through HTTPS and secure WebSockets (`wss://`).
- Configure the frontend with the deployed backend URL and provide test URLs reachable by that backend.
- Verify the platform's process management, private networking, resource limits and WebSocket support before finalizing the deployment layout.
- Keep test sources available during the review period, monitor resource usage and configure a spending limit.

Actual hosting cost must be measured. Continuous transcoding and video bandwidth are the main variables; avoid unnecessary source encoding and stop receiving FFmpeg processes when viewers disconnect.

## Implementation milestones

### 1. Prove the streaming pipeline

- Start MediaMTX through Docker.
- Publish one looping MP4 using FFmpeg.
- Verify the RTSP source is readable.
- Build a minimal Go WebSocket endpoint and browser decoder.
- Confirm live video appears and disconnecting cleans up FFmpeg.

### 2. Build the viewer interface

- Add URL entry, stream cards and a responsive grid.
- Add independent Play/Pause and Remove controls.
- Add connection status and readable error messages.
- Expand the initial single-source prototype to four distinct sources and verify all four play simultaneously in a 2 × 2 desktop grid.

### 3. Make the backend reliable

- Add validation, destination restrictions, stream limits and timeouts.
- Handle unavailable sources, FFmpeg exits and slow/disconnected clients.
- Verify repeated play/pause operations do not leak processes or memory.
- Add focused automated tests for validation and stream lifecycle behavior.

### 4. Deploy and document

- Deploy the frontend, backend and test-source setup.
- Verify the full pipeline from the public demo.
- Write the README and architecture explanation.
- Include demo URLs, supported controls, limitations and troubleshooting.

## Acceptance checklist

- A valid, reachable RTSP source plays in a browser.
- Four distinct test streams play together in a responsive grid.
- Each player can pause, resume and be removed independently.
- Invalid, offline and unauthorized sources produce clear errors.
- Disconnecting viewers releases their backend processes.
- Concurrent streams remain within configured resource limits.
- Local setup is reproducible with documented Docker commands.
- The live demo works over HTTPS and secure WebSockets.
- The repository and README satisfy the assignment deliverables.

## Decisions still to confirm

- Which four sample MP4 files to use; choose small files we have permission to distribute.
- Final decoder choice after the first streaming prototype.
- Maximum active streams after measuring performance; start with a limit of four.
- Hosting budget and how long the demo should remain online.
- Assignment receipt date and submission deadline.

These do not block starting the local streaming prototype.
