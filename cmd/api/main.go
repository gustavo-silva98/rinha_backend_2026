package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"rinha2026/internal/model"
	"rinha2026/internal/preprocess"
	"runtime"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
)

// var ctx = context.Background()
var payload model.Payload
var vetor model.Vector14Dim
var response model.Response

type Backend struct {
	Api_Id    int
	NormConst preprocess.NormalizationConstant
	MCCRisk   map[string]float64
}

func (api *Backend) TestEndpoint(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	fmt.Fprintf(w, "%v - %v", now, api.Api_Id)
}

func (api *Backend) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *Backend) FraudScore(w http.ResponseWriter, r *http.Request) {
	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Fatalf("Erro ao ler payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	api.Vetorize(&payload, &vetor)
	response.Approved = true
	response.FraudScore = 0
	jsonData, _ := sonic.Marshal(response)

	w.Write(jsonData)
}

func limitar(val float64) float64 {

	switch {
	case val > 1:
		return 1
	case val < 0:
		return 0
	}
	return val
}

func deParaWeekday(weekday time.Time) int {
	mapWeek := map[int]int{
		0: 6,
		1: 0,
		2: 1,
		3: 2,
		4: 3,
		5: 4,
		6: 5,
	}
	val := mapWeek[int(weekday.Weekday())]
	return val
}

func (api *Backend) Vetorize(payload *model.Payload, vetor *model.Vector14Dim) {
	requestedAt, _ := time.Parse(time.RFC3339, payload.Transaction.RequestedAt.String())
	timestampLastTrans, _ := time.Parse(time.RFC3339, payload.Transaction.RequestedAt.String())
	vetor.Dim0 = int8(limitar(payload.Transaction.Amount/float64(api.NormConst.MaxAmount)) * 127)
	vetor.Dim1 = int8(limitar(float64(payload.Transaction.Installments)/float64(api.NormConst.MaxInstallments)) * 127)
	vetor.Dim2 = int8(limitar((payload.Transaction.Amount/payload.Customer.AvgAmount)/float64(api.NormConst.AmountVsAvgRatio)) * 127)
	vetor.Dim3 = int8(requestedAt.Hour() / 23)
	vetor.Dim4 = int8(deParaWeekday(requestedAt) / 6)
	if timestampLastTrans.Year() < 2 {
		vetor.Dim5 = int8(-1)
		vetor.Dim6 = int8(-1)
	} else {
		vetor.Dim5 = int8(limitar(float64(timestampLastTrans.Minute())/float64(api.NormConst.MaxMinutes)) * 127)
		vetor.Dim6 = int8(limitar(payload.LastTransaction.KmFromCurrent / float64(api.NormConst.MaxKM)))
	}
	vetor.Dim7 = int8(limitar(payload.Terminal.KmFromHome / float64(api.NormConst.MaxKM)))
	vetor.Dim8 = int8(limitar(float64(payload.Customer.TxCount24H) / float64(api.NormConst.MaxTxCount24)))
	if payload.Terminal.IsOnline {
		vetor.Dim9 = int8(1)
	} else {
		vetor.Dim9 = int8(0)
	}

	if payload.Terminal.CardPresent {
		vetor.Dim10 = int8(1)
	} else {
		vetor.Dim10 = int8(0)
	}
	vetor.Dim11 = 1
	for _, val := range payload.Customer.KnownMerchants {
		if payload.Merchant.ID == val {
			vetor.Dim11 = 0
		}
	}
	val, ok := api.MCCRisk[payload.Merchant.Mcc]
	if !ok {
		vetor.Dim12 = int8(64)
	} else {
		vetor.Dim12 = int8((val) * 127)
	}
	vetor.Dim13 = int8(limitar(payload.Merchant.AvgAmount/float64(api.NormConst.MaxMerchantAvgAmount)) * 127)
}

func main() {
	if runtime.NumCPU()*4 > 32 {
		runtime.GOMAXPROCS(32)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
	}

	// Carregando variaveis de normalização e MCC
	normConst, err := preprocess.LoadNormalization()
	if err != nil {
		log.Fatalf("erro ao carregar Constantes de Normalização: %v", err)
		return
	}
	mcc, err := preprocess.LoadMCC()
	if err != nil {
		log.Fatalf("erro ao carregar MCC: %v", err)
		return
	}

	port := os.Getenv("PORT")
	api_id, _ := strconv.Atoi(os.Getenv("API_ID"))
	api := &Backend{
		Api_Id:    api_id,
		NormConst: normConst,
		MCCRisk:   mcc,
	}

	router := http.NewServeMux()
	router.HandleFunc("/test", api.TestEndpoint)
	router.HandleFunc("/ready", api.ReadyEndpoint)
	router.HandleFunc("/fraud-score", api.FraudScore)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}
