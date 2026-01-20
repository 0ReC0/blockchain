package financial

import (
	"encoding/xml"
	"fmt"
	"time"

	"blockchain/storage/txpool"
)

// ISO20022 структуры для финансовых сообщений
type Document struct {
	XMLName           xml.Name          `xml:"Document"`
	FIToFICstmrCdtTrf *FIToFICstmrCdtTrf `xml:"FIToFICstmrCdtTrf"`
}

type FIToFICstmrCdtTrf struct {
	GrpHdr     *GrpHdr        `xml:"GrpHdr"`
	CdtTrfTxInf []*CdtTrfTxInf `xml:"CdtTrfTxInf"`
}

type GrpHdr struct {
	MsgId   string `xml:"MsgId"`
	CreDtTm string `xml:"CreDtTm"`
	NbOfTxs string `xml:"NbOfTxs"`
	TtlIntrBkSttlmAmt *Amount `xml:"TtlIntrBkSttlmAmt"`
	IntrBkSttlmDt     string  `xml:"IntrBkSttlmDt"`
}

type Amount struct {
	Value    string `xml:",chardata"`
	Currency string `xml:"Ccy,attr"`
}

type CdtTrfTxInf struct {
	PmtId   *PaymentIdentification `xml:"PmtId"`
	PmtTpInf *PaymentTypeInformation `xml:"PmtTpInf"`
	Amt     *Amount                `xml:"Amt"`
	CdtrAgt *BranchAndFinancialInstitutionIdentification `xml:"CdtrAgt"`
	Cdtr    *PartyIdentification                         `xml:"Cdtr"`
	CdtrAcct *CashAccount                              `xml:"CdtrAcct"`
	Dbtr    *PartyIdentification                         `xml:"Dbtr"`
	DbtrAcct *CashAccount                              `xml:"DbtrAcct"`
	RmtInf  *RemittanceInformation                      `xml:"RmtInf"`
}

type PaymentIdentification struct {
	EndToEndId string `xml:"EndToEndId"`
	TxId       string `xml:"TxId"`
}

type PaymentTypeInformation struct {
	InstrPrty string `xml:"InstrPrty"`
	SvcLvl    *ServiceLevel `xml:"SvcLvl"`
	LclInstrm *LocalInstrument `xml:"LclInstrm"`
	CtgyPurp  *CategoryPurpose `xml:"CtgyPurp"`
}

type ServiceLevel struct {
	Cd string `xml:"Cd"`
}

type LocalInstrument struct {
	Cd string `xml:"Cd"`
}

type CategoryPurpose struct {
	Cd string `xml:"Cd"`
}

type BranchAndFinancialInstitutionIdentification struct {
	FinInstnId *FinancialInstitutionIdentification `xml:"FinInstnId"`
}

type FinancialInstitutionIdentification struct {
	BIC string `xml:"BIC"`
}

type PartyIdentification struct {
	Nm string `xml:"Nm"`
	PstlAdr *PostalAddress `xml:"PstlAdr"`
}

type PostalAddress struct {
	Ctry string `xml:"Ctry"`
	AdrLine []string `xml:"AdrLine"`
}

type CashAccount struct {
	Id *AccountIdentification `xml:"Id"`
}

type AccountIdentification struct {
	IBAN string `xml:"IBAN"`
	Othr *GenericAccountIdentification `xml:"Othr"`
}

type GenericAccountIdentification struct {
	Id string `xml:"Id"`
}

type RemittanceInformation struct {
	Ustrd string `xml:"Ustrd"`
}

// Адаптеры для конвертации
type ISO20022Adapter struct {
	InstitutionId string
}

type SWIFTAdapter struct {
	BIC string
}

// FinancialIntegration основная структура для интеграции с финансовыми системами
type FinancialIntegration struct {
	iso20022Adapter *ISO20022Adapter
	swiftAdapter    *SWIFTAdapter
}

// NewFinancialIntegration создает новый экземпляр финансовой интеграции
func NewFinancialIntegration(institutionId, bic string) *FinancialIntegration {
	return &FinancialIntegration{
		iso20022Adapter: &ISO20022Adapter{InstitutionId: institutionId},
		swiftAdapter:    &SWIFTAdapter{BIC: bic},
	}
}

// ConvertToISO20022 конвертирует транзакцию блокчейна в формат ISO 20022
func (f *FinancialIntegration) ConvertToISO20022(tx *txpool.Transaction) (*Document, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}

	doc := &Document{
		FIToFICstmrCdtTrf: &FIToFICstmrCdtTrf{
			GrpHdr: &GrpHdr{
				MsgId:   tx.ID,
				CreDtTm: time.Unix(tx.Timestamp, 0).Format(time.RFC3339),
				NbOfTxs: "1",
				TtlIntrBkSttlmAmt: &Amount{
					Value:    fmt.Sprintf("%.2f", tx.Amount),
					Currency: "USD", // Можно сделать параметризуемым
				},
				IntrBkSttlmDt: time.Unix(tx.Timestamp, 0).Format("2006-01-02"),
			},
			CdtTrfTxInf: []*CdtTrfTxInf{
				{
					PmtId: &PaymentIdentification{
						EndToEndId: tx.ID,
						TxId:       tx.ID,
					},
					PmtTpInf: &PaymentTypeInformation{
						InstrPrty: "NORM",
						SvcLvl:    &ServiceLevel{Cd: "SEPA"},
						LclInstrm: &LocalInstrument{Cd: "CORE"},
						CtgyPurp:  &CategoryPurpose{Cd: "CASH"},
					},
					Amt: &Amount{
						Value:    fmt.Sprintf("%.2f", tx.Amount),
						Currency: "USD",
					},
					CdtrAgt: &BranchAndFinancialInstitutionIdentification{
						FinInstnId: &FinancialInstitutionIdentification{
							BIC: f.swiftAdapter.BIC,
						},
					},
					Cdtr: &PartyIdentification{
						Nm: "Creditor Institution",
						PstlAdr: &PostalAddress{
							Ctry: "US",
							AdrLine: []string{"Creditor Address Line 1"},
						},
					},
					CdtrAcct: &CashAccount{
						Id: &AccountIdentification{
							Othr: &GenericAccountIdentification{
								Id: tx.To,
							},
						},
					},
					Dbtr: &PartyIdentification{
						Nm: "Debtor Institution",
						PstlAdr: &PostalAddress{
							Ctry: "US",
							AdrLine: []string{"Debtor Address Line 1"},
						},
					},
					DbtrAcct: &CashAccount{
						Id: &AccountIdentification{
							Othr: &GenericAccountIdentification{
								Id: tx.From,
							},
						},
					},
					RmtInf: &RemittanceInformation{
						Ustrd: fmt.Sprintf("Blockchain transaction %s", tx.ID),
					},
				},
			},
		},
	}

	return doc, nil
}

// ConvertBatchToISO20022 конвертирует группу транзакций в формат ISO 20022
func (f *FinancialIntegration) ConvertBatchToISO20022(transactions []*txpool.Transaction) (*Document, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions provided")
	}

	totalAmount := 0.0
	for _, tx := range transactions {
		totalAmount += tx.Amount
	}

	doc := &Document{
		FIToFICstmrCdtTrf: &FIToFICstmrCdtTrf{
			GrpHdr: &GrpHdr{
				MsgId:   fmt.Sprintf("batch-%d", time.Now().Unix()),
				CreDtTm: time.Now().Format(time.RFC3339),
				NbOfTxs: fmt.Sprintf("%d", len(transactions)),
				TtlIntrBkSttlmAmt: &Amount{
					Value:    fmt.Sprintf("%.2f", totalAmount),
					Currency: "USD",
				},
				IntrBkSttlmDt: time.Now().Format("2006-01-02"),
			},
			CdtTrfTxInf: make([]*CdtTrfTxInf, len(transactions)),
		},
	}

	for i, tx := range transactions {
		doc.FIToFICstmrCdtTrf.CdtTrfTxInf[i] = &CdtTrfTxInf{
			PmtId: &PaymentIdentification{
				EndToEndId: tx.ID,
				TxId:       tx.ID,
			},
			PmtTpInf: &PaymentTypeInformation{
				InstrPrty: "NORM",
				SvcLvl:    &ServiceLevel{Cd: "SEPA"},
				LclInstrm: &LocalInstrument{Cd: "CORE"},
				CtgyPurp:  &CategoryPurpose{Cd: "CASH"},
			},
			Amt: &Amount{
				Value:    fmt.Sprintf("%.2f", tx.Amount),
				Currency: "USD",
			},
			CdtrAgt: &BranchAndFinancialInstitutionIdentification{
				FinInstnId: &FinancialInstitutionIdentification{
					BIC: f.swiftAdapter.BIC,
				},
			},
			Cdtr: &PartyIdentification{
				Nm: "Creditor Institution",
				PstlAdr: &PostalAddress{
					Ctry: "US",
					AdrLine: []string{"Creditor Address Line 1"},
				},
			},
			CdtrAcct: &CashAccount{
				Id: &AccountIdentification{
					Othr: &GenericAccountIdentification{
						Id: tx.To,
					},
				},
			},
			Dbtr: &PartyIdentification{
				Nm: "Debtor Institution",
				PstlAdr: &PostalAddress{
					Ctry: "US",
					AdrLine: []string{"Debtor Address Line 1"},
				},
			},
			DbtrAcct: &CashAccount{
				Id: &AccountIdentification{
					Othr: &GenericAccountIdentification{
						Id: tx.From,
					},
				},
			},
			RmtInf: &RemittanceInformation{
				Ustrd: fmt.Sprintf("Blockchain transaction %s", tx.ID),
			},
		}
	}

	return doc, nil
}

// ToXML сериализует документ ISO20022 в XML
func (d *Document) ToXML() ([]byte, error) {
	return xml.MarshalIndent(d, "", "  ")
}

// FromXML десериализует XML в документ ISO20022
func (d *Document) FromXML(data []byte) error {
	return xml.Unmarshal(data, d)
}
