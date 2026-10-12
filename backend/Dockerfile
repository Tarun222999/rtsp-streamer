FROM golang:1.27-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 go build -o /out/server .

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg ca-certificates tini \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=build /out/server ./server
COPY public ./public

USER nobody:nogroup

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["tini", "--", "/app/server"]