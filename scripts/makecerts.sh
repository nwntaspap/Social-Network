#!/usr/bin/env bash

# Load variables from .env
set -a          # automatically export all variables
source .env
set +a

# Check if files already exist
if [[ -f "$SERVER_TLS_KEY_FILE"  ||  -f "$SERVER_TLS_CERT_FILE" ]]; then
  echo "Certificates already exist. Skipping generation."
  exit 0
fi


mkdir -p certs
openssl req -x509 \
  -newkey rsa:4096 \
  -sha256 \
  -nodes \
  -keyout ${SERVER_TLS_KEY_FILE} \
  -out ${SERVER_TLS_CERT_FILE} \
  -days 825 -subj "/CN=${SERVER_HOST}" \
  -addext "subjectAltName=DNS:${SERVER_HOST}"
echo "Certificate generated successfully."
