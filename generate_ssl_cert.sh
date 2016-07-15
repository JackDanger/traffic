#!/bin/bash

set -euo pipefail

key="server/key.pem"
cert="server/cert.pem"

# Generate an RSA private key
2>&1 openssl genrsa -des3 -out ${key}.locked -passout pass:insecure 1024 | head -n 1

# Remove the password from the RSA private key
openssl rsa -in ${key}.locked -out $key -passin pass:insecure 2>/dev/null

# Create a new CSR with the right subject
openssl req -new -key $key -out csr.pem -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com"

# Generate a certificate from the private key and the CSR
openssl req -x509 -days 365 -key $key -in csr.pem -out $cert

# Cleanup temp files
rm ${key}.locked
rm csr.pem

echo "A new private key/certificate pair has been generated here for you: $key/$cert"
