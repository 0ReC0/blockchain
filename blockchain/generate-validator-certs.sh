#!/bin/bash

# Папка для сертификатов
CERT_DIR="certs"
CA_KEY="$CERT_DIR/ca.key"
CA_CERT="$CERT_DIR/ca.crt"

# Количество валидаторов
VALIDATOR_COUNT=5

# Создаем папку (если не существует)
mkdir -p "$CERT_DIR"

# 1. Генерация корневого CA (RSA 4096)
if [ ! -f "$CA_KEY" ]; then
  echo "🔐 Создание корневого CA (RSA 4096)..."
  openssl genrsa -out "$CA_KEY" 4096
  openssl req -new -x509 -days 365 -key "$CA_KEY" -out "$CA_CERT" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=MyOrg/CN=mydomain.local"
fi

# 2. Генерация ключей и сертификатов для валидаторов (ECDSA)
echo "⛏️ Генерация ECDSA-ключей и сертификатов для $VALIDATOR_COUNT валидаторов..."

for i in $(seq 1 $VALIDATOR_COUNT); do
  NAME="validator$i"
  KEY="$CERT_DIR/${NAME}.key"
  CSR="$CERT_DIR/${NAME}.csr"
  CRT="$CERT_DIR/${NAME}.crt"

  # Генерация ECDSA-ключа (prime256v1)
  openssl ecparam -name prime256v1 -genkey -out "$KEY"

  # Создание CSR
  openssl req -new -key "$KEY" -out "$CSR" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=MyOrg/CN=$NAME"

  # Подписание сертификата через CA
  openssl x509 -req -days 365 -in "$CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$CRT"
done

# 3. Генерация серверного ключа и сертификата
SERVER_KEY="$CERT_DIR/server.key"
SERVER_CSR="$CERT_DIR/server.csr"
SERVER_CRT="$CERT_DIR/server.crt"

if [ ! -f "$SERVER_KEY" ]; then
  echo "🔧 Генерация серверного ключа и сертификата..."
  openssl ecparam -name prime256v1 -genkey -out "$SERVER_KEY"
  openssl req -new -key "$SERVER_KEY" -out "$SERVER_CSR" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=MyOrg/CN=server"
  openssl x509 -req -days 365 -in "$SERVER_CSR" -CA "$CA_CERT" -CAkey "$CA_KEY" -CAcreateserial -out "$SERVER_CRT"
fi

echo "✅ Все сертификаты успешно сгенерированы в папке: $CERT_DIR"