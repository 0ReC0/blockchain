package main

import (
	"fmt"
	"time"

	"blockchain/integration/financial"
	"blockchain/storage/txpool"
)

func main() {
	fmt.Println("=== Демонстрация интеграции блокчейна с финансовыми системами ===")

	// Создаем экземпляр финансовой интеграции
	// institutionId - идентификатор вашего финансового учреждения
	// BIC - банковский идентификационный код вашего банка
	fi := financial.NewFinancialIntegration("my-institution-123", "MYBANKBICXXX")

	// Создаем пример транзакции блокчейна
	tx := &txpool.Transaction{
		ID:        "bc-tx-20230101-001",
		From:      "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		To:        "12c6DSiU4Rq3P4ZxziKxzrL5LmMBrzjrJX",
		Amount:    1500.75,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "3045022100...", // Пример подписи
	}

	fmt.Printf("Исходная транзакция блокчейна:\n")
	fmt.Printf("  ID: %s\n", tx.ID)
	fmt.Printf("  Отправитель: %s\n", tx.From)
	fmt.Printf("  Получатель: %s\n", tx.To)
	fmt.Printf("  Сумма: %.2f\n", tx.Amount)
	fmt.Printf("  Комиссия: %.2f\n", tx.Fee)
	fmt.Printf("  Время: %s\n", time.Unix(tx.Timestamp, 0).Format("2006-01-02 15:04:05"))

	// Конвертируем транзакцию в формат ISO20022
	fmt.Println("\n--- Конвертация в формат ISO20022 ---")
	doc, err := fi.ConvertToISO20022(tx)
	if err != nil {
		fmt.Printf("Ошибка конвертации: %v\n", err)
		return
	}

	fmt.Printf("Транзакция успешно сконвертирована в формат ISO20022\n")
	fmt.Printf("Количество транзакций в сообщении: %s\n", doc.FIToFICstmrCdtTrf.GrpHdr.NbOfTxs)
	fmt.Printf("Общая сумма: %s %s\n",
		doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value,
		doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Currency)

	// Сериализуем в XML для отправки в финансовую систему
	fmt.Println("\n--- Сериализация в XML ---")
	xmlData, err := doc.ToXML()
	if err != nil {
		fmt.Printf("Ошибка сериализации в XML: %v\n", err)
		return
	}

	fmt.Printf("XML представление ISO20022 сообщения:\n%s\n", string(xmlData))

	// Демонстрация пакетной обработки
	fmt.Println("\n--- Пакетная обработка транзакций ---")
	batchTransactions := []*txpool.Transaction{
		{
			ID:        "bc-tx-20230101-002",
			From:      "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			To:        "12c6DSiU4Rq3P4ZxziKxzrL5LmMBrzjrJX",
			Amount:    2500.00,
			Fee:       0.01,
			Timestamp: time.Now().Unix(),
			Signature: "3045022100...",
		},
		{
			ID:        "bc-tx-20230101-003",
			From:      "12c6DSiU4Rq3P4ZxziKxzrL5LmMBrzjrJX",
			To:        "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			Amount:    750.25,
			Fee:       0.01,
			Timestamp: time.Now().Unix(),
			Signature: "3045022100...",
		},
	}

	// Конвертируем пакет транзакций
	batchDoc, err := fi.ConvertBatchToISO20022(batchTransactions)
	if err != nil {
		fmt.Printf("Ошибка конвертации пакета: %v\n", err)
		return
	}

	fmt.Printf("Пакет из %s транзакций успешно сконвертирован\n", batchDoc.FIToFICstmrCdtTrf.GrpHdr.NbOfTxs)
	fmt.Printf("Общая сумма пакета: %s %s\n",
		batchDoc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value,
		batchDoc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Currency)

	// Сериализуем пакет в XML
	batchXmlData, err := batchDoc.ToXML()
	if err != nil {
		fmt.Printf("Ошибка сериализации пакета в XML: %v\n", err)
		return
	}

	fmt.Printf("XML представление пакета ISO20022 сообщений:\n%s\n", string(batchXmlData))

	// Демонстрация десериализации
	fmt.Println("\n--- Десериализация XML обратно в объект ---")
	newDoc := &financial.Document{}
	err = newDoc.FromXML(xmlData)
	if err != nil {
		fmt.Printf("Ошибка десериализации XML: %v\n", err)
		return
	}

	fmt.Printf("Документ успешно восстановлен из XML\n")
	fmt.Printf("ID сообщения: %s\n", newDoc.FIToFICstmrCdtTrf.GrpHdr.MsgId)
	fmt.Printf("Время создания: %s\n", newDoc.FIToFICstmrCdtTrf.GrpHdr.CreDtTm)

	// Демонстрация конвертации в SEPA формат
	fmt.Println("\n--- Конвертация в формат SEPA ---")
	sepaDoc, err := fi.ConvertToSEPA(tx)
	if err != nil {
		fmt.Printf("Ошибка конвертации в SEPA: %v\n", err)
		return
	}

	fmt.Printf("Транзакция успешно сконвертирована в формат SEPA\n")
	fmt.Printf("Валюта: %s\n", sepaDoc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Currency)
	fmt.Printf("Сумма: %s\n", sepaDoc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value)

	// Сериализуем SEPA документ в XML
	sepaXmlData, err := sepaDoc.ToXML()
	if err != nil {
		fmt.Printf("Ошибка сериализации SEPA в XML: %v\n", err)
		return
	}

	fmt.Printf("XML представление SEPA сообщения:\n%s\n", string(sepaXmlData))

	fmt.Println("\n=== Интеграция завершена успешно ===")
	fmt.Println("Теперь вы можете отправить XML сообщения в вашу финансовую систему")
	fmt.Println("которая поддерживает стандарт ISO20022 или SEPA для обработки платежей.")
}
