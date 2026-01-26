package financial

import (
	"testing"
	"time"

	"blockchain/storage/txpool"
)

func BenchmarkConvertToISO20022(b *testing.B) {
	fi := NewFinancialIntegration("institution-123", "BIC12345")
	
	tx := &txpool.Transaction{
		ID:        "tx-001",
		From:      "sender-address-123",
		To:        "receiver-address-456",
		Amount:    100.50,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "test-signature",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fi.ConvertToISO20022(tx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConvertBatchToISO20022(b *testing.B) {
	fi := NewFinancialIntegration("institution-123", "BIC12345")

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fi.ConvertBatchToISO20022(transactions)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkToXML(b *testing.B) {
	fi := NewFinancialIntegration("institution-123", "BIC12345")

	tx := &txpool.Transaction{
		ID:        "tx-001",
		From:      "sender-address-123",
		To:        "receiver-address-456",
		Amount:    100.50,
		Fee:       0.01,
		Timestamp: time.Now().Unix(),
		Signature: "test-signature",
	}

	doc, err := fi.ConvertToISO20022(tx)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := doc.ToXML()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromXML(b *testing.B) {
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

	doc := &Document{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := doc.FromXML(xmlData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
