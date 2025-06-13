#!/bin/bash

set -e

# Папка для сертификатов
CERT_DIR="certs"
mkdir -p "$CERT_DIR"

# Временная папка для конфигов
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Количество валидаторов
VALIDATOR_COUNT=5

# Функция для создания openssl.cnf с SAN
create_openssl_cnf() {
    local cn="$1"
    cat > "$TMP_DIR/openssl-$cn.cnf" <<EOF
[ req ]
default_bits        = 2048
default_keyfile     = privkey.pem
distinguished_name  = req_distinguished_name
req_extensions      = req_ext
prompt              = no

[ req_distinguished_name ]
C  = RU
ST = Moscow
L  = Moscow
O  = MyOrg
CN = $cn

[ req_ext ]
subjectAltName = @alt_names
basicConstraints = CA:FALSE
keyUsage = digitalSignature

[ alt_names ]
DNS.1 = mydomain.local
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF
}

# Функция для создания openssl.cnf для CA
create_openssl_ca_cnf() {
    cat > "$TMP_DIR/openssl-ca.cnf" <<EOF
[ req ]
default_bits        = 2048
default_keyfile     = privkey.pem
distinguished_name  = req_distinguished_name
req_extensions      = req_ext
prompt              = no

[ req_distinguished_name ]
C  = RU
ST = Moscow
L  = Moscow
O  = MyOrg
CN = mydomain.local

[ req_ext ]
basicConstraints = CA:TRUE, pathlen:0
keyUsage = critical, keyCertSign, cRLSign
EOF
}

# Генерация EC-ключа для CA
echo "🔐 Generating EC private key for CA..."
openssl ecparam -name prime256v1 -genkey -out "$CERT_DIR/ca.key"

# Генерация CA-сертификата
echo "🔐 Generating CA certificate..."
create_openssl_ca_cnf
openssl req -new -key "$CERT_DIR/ca.key" -out "$TMP_DIR/ca.csr" \
  -config "$TMP_DIR/openssl-ca.cnf"
openssl x509 -req -days 365 \
  -in "$TMP_DIR/ca.csr" \
  -signkey "$CERT_DIR/ca.key" \
  -out "$CERT_DIR/ca.crt" \
  -extfile "$TMP_DIR/openssl-ca.cnf" -extensions req_ext \
  -sha256

# Генерация сертификатов для валидаторов
echo "🏷️ Generating validator certificates..."
for i in $(seq 1 $VALIDATOR_COUNT); do
    echo "🏷️ Generating certificate for validator$i..."
    create_openssl_cnf "validator$i"

    # Генерация EC-ключа
    openssl ecparam -name prime256v1 -genkey -out "$CERT_DIR/validator$i.key"

    # CSR
    openssl req -new -key "$CERT_DIR/validator$i.key" \
      -out "$TMP_DIR/validator$i.csr" \
      -config "$TMP_DIR/openssl-validator$i.cnf"

    # Подписанный сертификат
    openssl x509 -req -days 365 \
      -in "$TMP_DIR/validator$i.csr" \
      -CA "$CERT_DIR/ca.crt" \
      -CAkey "$CERT_DIR/ca.key" \
      -CAcreateserial \
      -out "$CERT_DIR/validator$i.crt" \
      -extfile "$TMP_DIR/openssl-validator$i.cnf" -extensions req_ext \
      -sha256
done

# Проверка
echo "🔍 Verifying certificates..."
for i in $(seq 1 $VALIDATOR_COUNT); do
    echo "📄 validator$i.crt SAN:"
    openssl x509 -in "$CERT_DIR/validator$i.crt" -text -noout | grep -A 2 "Subject Alternative Name"
done

echo "✅ All ECDSA certificates generated successfully in '$CERT_DIR'"