# BUILD STEP
FROM golang:1.12.0-stretch AS builder
RUN mkdir -p /root/gomitmproxy
WORKDIR /root/gomitmproxy
COPY . /root/gomitmproxy/
RUN make build

# BUILDING SELENIUM
FROM selenium/standalone-chrome-debug:3.141.59

COPY selenium/gomitmproxy_entrypoint.sh  /opt/bin/gomitmproxy_entrypoint.sh
COPY selenium/gomitmproxy.conf /etc/supervisor/conf.d/
COPY selenium/config.json /etc/gomitmproxy/config.json
COPY selenium/start_gomitmproxy.sh /opt/bin/start_gomitmproxy.sh
COPY --from=builder /root/gomitmproxy/gomitmproxy /opt/bin/gomitmproxy

RUN sudo apt-get update && sudo apt-get install libnss3-tools -y

CMD ["/opt/bin/gomitmproxy_entrypoint.sh"]