# 📚 Имитационная модель блокчейна

Это учебная имитационная модель блокчейна с поддержкой нескольких алгоритмов консенсуса (PoS, BFT), шардинга, ZKP, смарт-контрактов и других функций.

---

## 🧰 Требования

- **Go 1.21+**
- **OpenSSL** (для генерации TLS-сертификатов)

---

## 🚀 Быстрый старт

### 1. Клонируйте репозиторий

```bash
git clone <ваш_репозиторий>
cd Blockchain
```

### 2. Сгенерируйте TLS-сертификаты

Система использует TLS для безопасного P2P-соединения. Необходимо сгенерировать сертификаты:

```bash
cd blockchain
mkdir -p certs
openssl genrsa -out certs/ca.key 4096
openssl req -new -x509 -days 365 -key certs/ca.key -out certs/ca.crt -subj "/CN=Test CA"
openssl genrsa -out certs/server.key 4096
openssl req -new -key certs/server.key -out certs/server.csr -subj "/CN=localhost"
openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt -days 365
```

### 3. Запустите систему

```bash
go run main.go
```

---

## 🌐 Доступные точки

- **REST API**: `http://localhost:8081`
  - `GET /blocks` — получить все блоки
  - `GET /transactions` — получить транзакции из пула

---

## 🧪 Примеры использования API

### Получить список блоков

```bash
curl http://localhost:8081/blocks
```

### Получить список транзакций

```bash
curl http://localhost:8081/transactions
```

---

## 🛠️ Дополнительно

### Использование Makefile (опционально)

Вы можете добавить `Makefile` для упрощения запуска:

```makefile
certs:
	mkdir -p certs
	openssl genrsa -out certs/ca.key 4096
	openssl req -new -x509 -days 365 -key certs/ca.key -out certs/ca.crt -subj "/CN=Test CA"
	openssl genrsa -out certs/server.key 4096
	openssl req -new -key certs/server.key -out certs/server.csr -subj "/CN=localhost"
	openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt -days 365

run: certs
	go run main.go
```

Запуск:

```bash
make run
```

---

## 📌 Важно

- Если вы хотите использовать **множество узлов**, необходимо сгенерировать отдельные сертификаты для каждого.
- Для тестовой среды можно отключить строгую проверку TLS, но это **не рекомендуется** для продакшена.

---

## 📚 Архитектура системы

См. диаграмму UML в репозитории: `architecture.puml`

---

## 🧑‍💻 Разработка

- **main.go** — точка входа
- **consensus/** — реализация консенсусов (PoS, BFT)
- **network/** — P2P-сеть, Gossip-протокол
- **storage/** — блокчейн, пул транзакций
- **crypto/** — криптография (подписи, хеши)
- **governance/** — голосование, обновления
- **security/** — защита от атак (Sybil, double spend)
- **integration/** — API, банковский шлюз
- **privacy/** — приватность (ZKP, гомоморфное шифрование)
- **scalability/** — шардинг, rollups

---

## 📬 Связь

Если у вас есть вопросы или предложения — пишите в разделе **Issues**.