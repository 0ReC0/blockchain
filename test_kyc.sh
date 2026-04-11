#!/bin/bash

echo "🧪 Тестирование KYC в клиенте"
echo "=============================="
echo ""

API="http://localhost:8081"
CLIENT="http://localhost:8000"
USER="testuser_$(date +%s)"

echo "1️⃣  KYC Регистрация через клиент..."
REG_RESULT=$(curl -s -X POST "$CLIENT/kyc/register" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\",\"fullName\":\"Test User\",\"idNumber\":\"ID123456\",\"country\":\"US\"}")

if [[ "$REG_RESULT" == *"KYC регистрация успешна"* ]]; then
    echo "✅ Успешно: $REG_RESULT"
else
    echo "❌ Ошибка: $REG_RESULT"
fi
echo ""

echo "2️⃣  KYC Верификация через клиент..."
VERIFY_RESULT=$(curl -s -X POST "$CLIENT/kyc/verify" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\"}")

if [[ "$VERIFY_RESULT" == *"KYC верификация успешна"* ]]; then
    echo "✅ Успешно: $VERIFY_RESULT"
else
    echo "❌ Ошибка: $VERIFY_RESULT"
fi
echo ""

echo "3️⃣  Проверка статуса KYC..."
STATUS_RESULT=$(curl -s "$CLIENT/kyc/status/$USER")

if [[ "$STATUS_RESULT" == *"Verified"* ]]; then
    echo "✅ Статус: $STATUS_RESULT"
else
    echo "⚠️  Статус: $STATUS_RESULT"
fi
echo ""

echo "4️⃣  Проверка на сервере..."
SERVER_STATUS=$(curl -s "$API/kyc/verify" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\"}")
echo "   Сервер: $SERVER_STATUS"
echo ""

echo "=============================="
echo "✅ Тест завершен!"
