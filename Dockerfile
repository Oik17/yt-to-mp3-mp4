FROM golang:1.23 AS builder

WORKDIR /app
COPY . .
RUN go build -o yt-downloader

FROM debian:bookworm-slim

# Install dependencies and pipx
RUN apt-get update && \
    apt-get install -y ffmpeg python3 python3-pip python3-venv pipx curl && \
    pipx install yt-dlp && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

ENV PATH="/root/.local/bin:$PATH"

WORKDIR /root/
COPY --from=builder /app/yt-downloader .
EXPOSE 8080

CMD ["./yt-downloader"]
