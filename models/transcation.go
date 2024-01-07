package models

type Transaction struct {
	Step          int
	Customer      string
	ZipCodeOrigin string
	Merchant      string
	ZipMerchant   string
	Age           int
	Gender        string
	Category      string
	Amount        float64
	Fraud         bool
}
