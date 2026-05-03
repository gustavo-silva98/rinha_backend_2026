package model

import "time"

type Payload struct {
	ID          string `json:"id"`
	Transaction struct {
		Amount       float64   `json:"amount"`
		Installments int       `json:"installments"`
		RequestedAt  time.Time `json:"requested_at"`
	} `json:"transaction"`
	Customer struct {
		AvgAmount      float64  `json:"avg_amount"`
		TxCount24H     int      `json:"tx_count_24h"`
		KnownMerchants []string `json:"known_merchants"`
	} `json:"customer"`
	Merchant struct {
		ID        string  `json:"id"`
		Mcc       string  `json:"mcc"`
		AvgAmount float64 `json:"avg_amount"`
	} `json:"merchant"`
	Terminal struct {
		IsOnline    bool    `json:"is_online"`
		CardPresent bool    `json:"card_present"`
		KmFromHome  float64 `json:"km_from_home"`
	} `json:"terminal"`
	LastTransaction struct {
		Timestamp     time.Time `json:"timestamp"`
		KmFromCurrent float64   `json:"km_from_current"`
	} `json:"last_transaction"`
}

type Vector14Dim struct {
	Dim0  int8
	Dim1  int8
	Dim2  int8
	Dim3  int8
	Dim4  int8
	Dim5  int8
	Dim6  int8
	Dim7  int8
	Dim8  int8
	Dim9  int8
	Dim10 int8
	Dim11 int8
	Dim12 int8
	Dim13 int8
}
