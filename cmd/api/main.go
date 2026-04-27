package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//var ctx = context.Background()

type Backend struct {
	Api_Id int
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

func main() {

	port := os.Getenv("PORT")
	api_id, _ := strconv.Atoi(os.Getenv("API_ID"))
	api := &Backend{
		Api_Id: api_id,
	}

	router := http.NewServeMux()
	router.HandleFunc("/test", api.TestEndpoint)
	router.HandleFunc("/ready", api.ReadyEndpoint)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}
