#!/bin/bash

# Тест регистрации с реальным ключом

echo "🧪 Тест регистрации публичного ключа"
echo "===================================="
echo ""

# 1. Запуск сервера
echo "1️⃣  Запуск сервера..."
cd blockchain
./blockchain-node > /tmp/node.log 2>&1 &
NODE_PID=$!
sleep 4

if ! kill -0 $NODE_PID 2>/dev/null; then
    echo "❌ Сервер не запустился"
    exit 1
fi
echo "✅ Сервер запущен (PID: $NODE_PID)"
echo ""

# 2. Генерация ключа
echo "2️⃣  Генерация ключа..."
cd ..
PUB_KEY=$(go run generate_keys.go 2>&1 | grep "04" | grep -v "💡" | head -1 | xargs)
USER="testuser_$(date +%s)"

if [ -z "$PUB_KEY" ]; then
    echo "❌ Не удалось сгенерировать ключ"
    kill $NODE_PID
    exit 1
fi

echo "   Пользователь: $USER"
echo "   Публичный ключ: ${PUB_KEY:0:50}..."
echo ""

# 3. Регистрация ключа
echo "3️⃣  Регистрация публичного ключа..."
RESULT=$(curl -s -X POST "http://localhost:8081/register" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\",\"pubKey\":\"$PUB_KEY\"}")

if [[ "$RESULT" == *"Public key registered"* ]]; then
    echo "✅ Успешно: $RESULT"
else
    echo "❌ Ошибка: $RESULT"
    kill $NODE_PID
    exit 1
fi
echo ""

# 4. KYC регистрация
echo "4️⃣  KYC регистрация..."
KYC_RESULT=$(curl -s -X POST "http://localhost:8081/kyc/register" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\",\"fullName\":\"Test User\",\"idNumber\":\"ID123\",\"country\":\"US\"}")

if [[ "$KYC_RESULT" == *"initiated"* ]]; then
    echo "✅ Успешно: $KYC_RESULT"
else
    echo "❌ Ошибка: $KYC_RESULT"
fi
echo ""

# 5. KYC верификация
echo "5️⃣  KYC верификация..."
VERIFY_RESULT=$(curl -s -X POST "http://localhost:8081/kyc/verify" \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"$USER\"}")

if [[ "$VERIFY_RESULT" == *"verified"* ]]; then
    echo "✅ Успешно: $VERIFY_RESULT"
else
    echo "❌ Ошибка: $VERIFY_RESULT"
fi
echo ""

# 6. Проверка блоков
echo "6️⃣  Проверка блоков..."
BLOCKS=$(curl -s "http://localhost:8081/blocks")
if [ -n "$BLOCKS" ]; then
    echo "✅ Блоки доступны: $BLOCKS" | head -100
else
    echo "❌ Блоки недоступны"
fi
echo ""

# Остановка
echo "🛑 Остановка сервера..."
kill $NODE_PID 2>/dev/null
echo ""
echo "===================================="
echo "✅ Тест завершен успешно!"
