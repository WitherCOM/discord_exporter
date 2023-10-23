FROM golang:alpine as builder
WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o discord_exporter .

FROM alpine:latest
WORKDIR /
COPY --from=builder /build/discord_exporter /discord_exporter
EXPOSE 9101
USER nonroot:nonroot
ENTRYPOINT ["/discord_exporter"]

