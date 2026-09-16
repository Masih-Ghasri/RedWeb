package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

type PaymentResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("POST /pay", func(w http.ResponseWriter, r *http.Request) {
		var req PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Simulate network latency between 100ms and 500ms
		time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")

		trxID := req.OrderID
		if len(trxID) > 8 {
			trxID = trxID[:8]
		}

		// Simulate 80% success and 20% transient gateway failure
		if rand.Float32() < 0.8 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PaymentResponse{
				Success:       true,
				TransactionID: "TRX-" + trxID,
				Message:       "Payment successful",
			})
			log.Printf("Payment SUCCESS for order: %s", req.OrderID)
		} else {
			w.WriteHeader(http.StatusPaymentRequired)
			json.NewEncoder(w).Encode(PaymentResponse{
				Success:       false,
				TransactionID: "",
				Message:       "Insufficient funds or bank timeout",
			})
			log.Printf("Payment FAILED for order: %s", req.OrderID)
		}
	})

	log.Println("Mock Bank API running on port 9090...")
	log.Fatal(http.ListenAndServe(":9090", mux))
}
