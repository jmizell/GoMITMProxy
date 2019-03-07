#!/usr/bin/env bash
CERT_FILE="/etc/gomitmproxy/ca.crt"
KEY_FILE="/etc/gomitmproxy/ca.key"
CERT_NAME="GoMITMProxy"

if ! [[ -f ${CERT_FILE} ]] || ! [[ -f ${KEY_FILE} ]]; then
    sudo opt/bin/gomitmproxy -generate_ca_only -key_age_hours 48 -ca_key_file ${KEY_FILE} -ca_cert_file ${CERT_FILE} \
    || exit 1
    sudo chmod 644 ${CERT_FILE}
fi

google-chrome --headless --disable-gpu --dump-dom https://www.chromestatus.com/ >/dev/null

for certDB in $(find ~/ -name "cert8.db"); do
    echo "Install ${CERT_NAME} ${CERT_FILE} to ${certDB}"
    if ! certutil -A -n "${CERT_NAME}" -t "TCu,Cu,Tu" -i ${CERT_FILE} -d "dbm:$(dirname ${certDB})"; then
        echo "!! Failed to install cert into ${certDB}"
        exit 1
    fi
done

for certDB in $(find ~/ -name "cert9.db"); do
    echo "Install ${CERT_NAME} ${CERT_FILE} to ${certDB}"
    if ! certutil -A -n "${CERT_NAME}" -t "TCu,Cu,Tu" -i ${CERT_FILE} -d "sql:$(dirname ${certDB})"; then
        echo "!! Failed to install cert into ${certDB}"
        exit 1
    fi
done

echo "nameserver 127.0.0.50" | sudo tee /etc/resolv.conf

/opt/bin/entry_point.sh