# BUILD STEP
FROM golang:1.12.0-stretch AS builder
RUN mkdir -p /root/gomitmproxy
WORKDIR /root/gomitmproxy
COPY . /root/gomitmproxy/
RUN make build

# BUILDING APP
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /root/gomitmproxy/gomitmproxy /root/gomitmproxy
ENTRYPOINT ["/root/gomitmproxy"]