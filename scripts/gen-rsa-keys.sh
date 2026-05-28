#!/bin/bash
set -euo pipefail

OUT_DIR="${1:-./configs}"

echo "Generating RSA 2048 key pair in ${OUT_DIR}..."

openssl genrsa -out "${OUT_DIR}/jwt-private.pem" 2048
openssl rsa -in "${OUT_DIR}/jwt-private.pem" -pubout -out "${OUT_DIR}/jwt-public.pem"

chmod 600 "${OUT_DIR}/jwt-private.pem"

echo "Done:"
echo "  Private key: ${OUT_DIR}/jwt-private.pem"
echo "  Public key:  ${OUT_DIR}/jwt-public.pem"
echo ""
echo "Copy jwt-public.pem to gateway/configs/ for verification"
