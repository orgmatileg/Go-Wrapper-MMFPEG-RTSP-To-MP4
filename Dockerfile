FROM golang:alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

FROM alpine:latest as runner

RUN apk add  --no-cache ffmpeg

WORKDIR /opt/go-wrapper-ffmpeg-rtsp-to-mp4/

COPY --from=builder /app/app .
COPY --from=builder /app/config.yaml config.yaml

ENTRYPOINT [ "/opt/go-wrapper-ffmpeg-rtsp-to-mp4/app"]
