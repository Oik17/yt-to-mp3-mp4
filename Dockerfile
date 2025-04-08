FROM golang:1.23 as builder

WORKDIR /app
COPY . .
RUN go build -o yt-downloader

FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y ffmpeg python3 python3-pip curl && \
    pip3 install yt-dlp && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/yt-downloader .
EXPOSE 8080

CMD ["./yt-downloader"]
