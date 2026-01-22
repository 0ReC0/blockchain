package financial

import (
	"testing"
	"time"

	"blockchain/storage/txpool"
)

func TestConvertToISO20022(t *testing.T) {
	// Создаем экземпляр финансовой интеграции
	fi := NewFinancialIntegration("institution-123", "BIC12345")

	// Создаем тестовую транзакцию
	tx := &txpool.Transaction{
		ID:        "tx-001",
		From:      "sender-address-123",
		To:        "receiver-address-456",
		Amount:    100.50,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "test-signature",
	}

	// Конвертируем транзакцию в формат ISO20022
	doc, err := fi.ConvertToISO20022(tx)
	if err != nil {
		t.Fatalf("Failed to convert transaction to ISO20022: %v", err)
	}

	// Проверяем результаты
	if doc.FIToFICstmrCdtTrf == nil {
		t.Error("FIToFICstmrCdtTrf should not be nil")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr == nil {
		t.Error("Group header should not be nil")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr.MsgId != tx.ID {
		t.Errorf("Expected MsgId %s, got %s", tx.ID, doc.FIToFICstmrCdtTrf.GrpHdr.MsgId)
	}

	if len(doc.FIToFICstmrCdtTrf.CdtTrfTxInf) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(doc.FIToFICstmrCdtTrf.CdtTrfTxInf))
	}

	// Проверяем сумму
	expectedAmount := "100.50"
	if doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value != expectedAmount {
		t.Errorf("Expected amount %s, got %s", expectedAmount, doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value)
	}
}

func TestConvertBatchToISO20022(t *testing.T) {
	// Создаем экземпляр финансовой интеграции
	fi := NewFinancialIntegration("institution-123", "BIC12345")

	// Создаем тестовые транзакции
	transactions := []*txpool.Transaction{
		{
			ID:        "tx-001",
			From:      "sender-1",
			To:        "receiver-1",
			Amount:    100.00,
			Fee:       0.01,
			Timestamp: time.Now().Unix(),
			Signature: "sig-1",
		},
		{
			ID:        "tx-002",
			From:      "sender-2",
			To:        "receiver-2",
			Amount:    200.00,
			Fee:       0.01,
			Timestamp: time.Now().Unix(),
			Signature: "sig-2",
		},
	}

	// Конвертируем транзакции в формат ISO20022
	doc, err := fi.ConvertBatchToISO20022(transactions)
	if err != nil {
		t.Fatalf("Failed to convert batch to ISO20022: %v", err)
	}

	// Проверяем результаты
	if doc.FIToFICstmrCdtTrf == nil {
		t.Error("FIToFICstmrCdtTrf should not be nil")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr == nil {
		t.Error("Group header should not be nil")
	}

	expectedCount := "2"
	if doc.FIToFICstmrCdtTrf.GrpHdr.NbOfTxs != expectedCount {
		t.Errorf("Expected %s transactions, got %s", expectedCount, doc.FIToFICstmrCdtTrf.GrpHdr.NbOfTxs)
	}

	expectedTotal := "300.00"
	if doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value != expectedTotal {
		t.Errorf("Expected total amount %s, got %s", expectedTotal, doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value)
	}

	if len(doc.FIToFICstmrCdtTrf.CdtTrfTxInf) != len(transactions) {
		t.Errorf("Expected %d transactions, got %d", len(transactions), len(doc.FIToFICstmrCdtTrf.CdtTrfTxInf))
	}
}

func TestToXML(t *testing.T) {
	// Создаем экземпляр финансовой интеграции
	fi := NewFinancialIntegration("institution-123", "BIC12345")

	// Создаем тестовую транзакцию
	tx := &txpool.Transaction{
		ID:        "tx-001",
		From:      "sender-address-123",
		To:        "receiver-address-456",
		Amount:    100.50,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "test-signature",
	}

	// Конвертируем транзакцию в формат ISO20022
	doc, err := fi.ConvertToISO20022(tx)
	if err != nil {
		t.Fatalf("Failed to convert transaction to ISO20022: %v", err)
	}

	// Сериализуем в XML
	xmlData, err := doc.ToXML()
	if err != nil {
		t.Fatalf("Failed to serialize to XML: %v", err)
	}

	// Проверяем, что XML не пустой
	if len(xmlData) == 0 {
		t.Error("XML data should not be empty")
	}

	// Проверяем, что XML содержит ожидаемые элементы
	xmlString := string(xmlData)
	if !contains(xmlString, "Document") {
		t.Error("XML should contain Document element")
	}
	if !contains(xmlString, "FIToFICstmrCdtTrf") {
		t.Error("XML should contain FIToFICstmrCdtTrf element")
	}
	if !contains(xmlString, tx.ID) {
		t.Error("XML should contain transaction ID")
	}
}

func TestFromXML(t *testing.T) {
	// Создаем тестовый XML документ
	xmlData := []byte(`<Document>
  <FIToFICstmrCdtTrf>
    <GrpHdr>
      <MsgId>test-tx-123</MsgId>
      <CreDtTm>2023-01-01T10:00:00Z</CreDtTm>
      <NbOfTxs>1</NbOfTxs>
      <TtlIntrBkSttlmAmt Ccy="USD">100.50</TtlIntrBkSttlmAmt>
      <IntrBkSttlmDt>2023-01-01</IntrBkSttlmDt>
    </GrpHdr>
    <CdtTrfTxInf>
      <PmtId>
        <EndToEndId>test-tx-123</EndToEndId>
        <TxId>test-tx-123</TxId>
      </PmtId>
      <Amt>
        <InstdAmt Ccy="USD">100.50</InstdAmt>
      </Amt>
    </CdtTrfTxInf>
  </FIToFICstmrCdtTrf>
</Document>`)

	// Создаем документ и десериализуем XML
	doc := &Document{}
	err := doc.FromXML(xmlData)
	if err != nil {
		t.Fatalf("Failed to deserialize from XML: %v", err)
	}

	// Проверяем результаты
	if doc.FIToFICstmrCdtTrf == nil {
		t.Error("FIToFICstmrCdtTrf should not be nil after deserialization")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr == nil {
		t.Error("Group header should not be nil after deserialization")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr.MsgId != "test-tx-123" {
		t.Errorf("Expected MsgId test-tx-123, got %s", doc.FIToFICstmrCdtTrf.GrpHdr.MsgId)
	}
}

func TestConvertToSEPA(t *testing.T) {
	// Создаем экземпляр финансовой интеграции
	fi := NewFinancialIntegration("institution-123", "BIC12345")

	// Создаем тестовую транзакцию
	tx := &txpool.Transaction{
		ID:        "tx-001",
		From:      "DE44500800000000000000", // Пример IBAN
		To:        "DE55500800000000000000", // Пример IBAN
		Amount:    100.50,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "test-signature",
	}

	// Конвертируем транзакцию в формат SEPA
	doc, err := fi.ConvertToSEPA(tx)
	if err != nil {
		t.Fatalf("Failed to convert transaction to SEPA: %v", err)
	}

	// Проверяем результаты
	if doc.FIToFICstmrCdtTrf == nil {
		t.Error("FIToFICstmrCdtTrf should not be nil")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr == nil {
		t.Error("Group header should not be nil")
	}

	if doc.FIToFICstmrCdtTrf.GrpHdr.MsgId != tx.ID {
		t.Errorf("Expected MsgId %s, got %s", tx.ID, doc.FIToFICstmrCdtTrf.GrpHdr.MsgId)
	}

	// Проверяем валюту (должна быть EUR для SEPA)
	if doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Currency != "EUR" {
		t.Errorf("Expected currency EUR, got %s", doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Currency)
	}

	// Проверяем сумму
	expectedAmount := "100.50"
	if doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value != expectedAmount {
		t.Errorf("Expected amount %s, got %s", expectedAmount, doc.FIToFICstmrCdtTrf.GrpHdr.TtlIntrBkSttlmAmt.Value)
	}

	// Проверяем, что используется IBAN
	if len(doc.FIToFICstmrCdtTrf.CdtTrfTxInf) > 0 {
		if doc.FIToFICstmrCdtTrf.CdtTrfTxInf[0].CdtrAcct.Id.IBAN == "" {
			t.Error("Creditor account should have IBAN for SEPA")
		}
		if doc.FIToFICstmrCdtTrf.CdtTrfTxInf[0].DbtrAcct.Id.IBAN == "" {
			t.Error("Debtor account should have IBAN for SEPA")
		}
	}
}

// Вспомогательная функция для проверки содержания строки
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

// Вспомогательная функция для поиска подстроки
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
