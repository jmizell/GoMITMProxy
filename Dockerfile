FROM golang:1.12.0-stretch
RUN mkdir -p /root/gomitmproxy
WORKDIR /root/gomitmproxy
COPY . /root/gomitmproxy/
RUN make

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /root/gomitmproxy/gomitmproxy /root/gomitmproxy
CMD ["/root/gomitmproxy"]