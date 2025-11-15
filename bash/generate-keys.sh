#!/bin/bash

mkdir -p keys

# Generate private key
openssl genrsa -out keys/private.pem 4096

# Generate public key
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

echo "RSA keys generated in keys/ directory"