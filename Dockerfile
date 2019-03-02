FROM golang:1.12.0-stretch
RUN mkdir -p /root/gomitmproxy
WORKDIR /root/gomitmproxy
COPY cmd /root/gomitmproxy/cmd
COPY proxy /root/gomitmproxy/proxy
COPY go.mod /root/gomitmproxy/go.mod
COPY go.sum /root/gomitmproxy/go.sum
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
        -a \
        -installsuffix cgo \
        -o gomitmproxy \
        github.com/jmizell/GoMITMProxy/cmd/gomitmproxy

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /root/gomitmproxy/gomitmproxy /root/gomitmproxy
CMD ["/root/gomitmproxy"]