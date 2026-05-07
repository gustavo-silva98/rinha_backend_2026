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
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

// var ctx = context.Background()

var ready bool
var payloadPool = sync.Pool{
	New: func() any {
		return &model.Payload{}
	},
}

var vetorPool = sync.Pool{
	New: func() any {
		return &model.Vector14Dim{}
	},
}

var responsePool = sync.Pool{
	New: func() any {
		return &model.Response{}
	},
}

type Backend struct {
	Api_Id    int
	NormConst preprocess.NormalizationConstant
	MCCRisk   map[string]float64
	Refs      *preprocess.StoredRefs
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
	if ready {
		w.WriteHeader(http.StatusOK)
	}
}

func (api *Backend) FraudScore(w http.ResponseWriter, r *http.Request) {
	payloadPtr := payloadPool.Get().(*model.Payload)
	defer payloadPool.Put(payloadPtr)
	*payloadPtr = model.Payload{}

	vetorPtr := vetorPool.Get().(*model.Vector14Dim)
	defer vetorPool.Put(vetorPtr)
	*vetorPtr = model.Vector14Dim{}

	responsePtr := responsePool.Get().(*model.Response)
	defer responsePool.Put(responsePtr)
	*responsePtr = model.Response{}

	err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(payloadPtr)
	if err != nil {
		log.Fatalf("Erro ao ler payload: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	api.Vetorize(payloadPtr, vetorPtr)
	responsePtr.Approved = api.Refs.ConvertLabel(1)
	responsePtr.FraudScore = 0
	jsonData, _ := sonic.Marshal(responsePtr)

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
func warmupSonic() error {
	const warmupJSON = `{
    "id": "tx-3943816664",
    "transaction": {
      "amount": 6384.53,
      "installments": 8,
      "requested_at": "2026-03-10T05:22:29Z"
    },
    "customer": {
      "avg_amount": 108.4,
      "tx_count_24h": 20,
      "known_merchants": [
        "MERC-010",
        "MERC-005",
        "MERC-001",
        "MERC-015"
      ]
    },
    "merchant": {
      "id": "MERC-074",
      "mcc": "7802",
      "avg_amount": 50.11
    },
    "terminal": {
      "is_online": true,
      "card_present": false,
      "km_from_home": 271.4091990309
    },
    "last_transaction": {
      "timestamp": "2026-03-10T05:21:29Z",
      "km_from_current": 550.307568803
    }
  }`

	payloadPtr := payloadPool.Get().(*model.Payload)
	defer payloadPool.Put(payloadPtr)
	*payloadPtr = model.Payload{}

	responsePtr := responsePool.Get().(*model.Response)
	defer responsePool.Put(responsePtr)
	*responsePtr = model.Response{}
	responsePtr.Approved = true
	responsePtr.FraudScore = 0
	for i := 1; i < 10; i++ {

		err := sonic.Unmarshal([]byte(warmupJSON), payloadPtr)
		if err != nil {
			return err
		}
		_, err = sonic.Marshal(responsePtr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (api *Backend) Vetorize(payload *model.Payload, vetor *model.Vector14Dim) {
	vetor.Dim0 = int8(limitar(payload.Transaction.Amount/float64(api.NormConst.MaxAmount)) * 127)
	vetor.Dim1 = int8(limitar(float64(payload.Transaction.Installments)/float64(api.NormConst.MaxInstallments)) * 127)
	vetor.Dim2 = int8(limitar((payload.Transaction.Amount/payload.Customer.AvgAmount)/float64(api.NormConst.AmountVsAvgRatio)) * 127)
	vetor.Dim3 = int8(payload.Transaction.RequestedAt.Hour() / 23)
	vetor.Dim4 = int8(deParaWeekday(payload.Transaction.RequestedAt) / 6)
	if payload.LastTransaction.Timestamp.Year() < 2 {
		vetor.Dim5 = int8(-1)
		vetor.Dim6 = int8(-1)
	} else {
		vetor.Dim5 = int8(limitar(float64(payload.LastTransaction.Timestamp.Minute())/float64(api.NormConst.MaxMinutes)) * 127)
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
			break
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
	// Carregando variaveis de normalização e MCC
	runtime.GOMAXPROCS(1)
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
	err = warmupSonic()
	if err != nil {
		log.Fatalf("erro ao WarmUp Sonic Json: %v", err)
		return
	}
	refs, err := preprocess.LoadRefs("vectors.bin")
	if err != nil {
		log.Fatalf("erro ao carregar as refs: %v", err)
	}

	port := os.Getenv("PORT")
	api_id, _ := strconv.Atoi(os.Getenv("API_ID"))
	api := &Backend{
		Api_Id:    api_id,
		NormConst: normConst,
		MCCRisk:   mcc,
		Refs:      refs,
	}
	ready = true

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
