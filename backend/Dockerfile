FROM golang:1.25-bookworm AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl ffmpeg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/server /app/server

ENV PORT=8080
ENV SQLITE_PATH=/app/data/podcast.db
ENV LOCAL_STORAGE_PATH=/app/data/storage
ENV FFMPEG_PATH=ffmpeg
EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=3s --retries=12 --start-period=5s \
    CMD curl -fsS http://127.0.0.1:8080/health || exit 1

CMD ["/app/server"]
