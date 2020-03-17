FROM golang:alpine AS builder

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"
CMD ["./piot-adapter-vdgps"]

FROM alpine:latest AS alpine
COPY --from=builder /app/piot-adapter-vdgps /app/piot-adapter-vdgps
WORKDIR /app/
EXPOSE 8080
CMD ["./piot-adapter-vdgps"]
