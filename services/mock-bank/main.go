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

	mux.HandleFunc("POST /pay", func(w http.ResponseWriter, r *http.Request) {
		var req PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// شبیه‌سازی تاخیر شبکه (Latancy) بین 100 تا 500 میلی‌ثانیه
		time.Sleep(time.Duration(rand.Intn(400)+100) * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")

		// شبیه‌سازی 80% موفقیت و 20% خطای درگاه بانکی
		if rand.Float32() < 0.8 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PaymentResponse{
				Success:       true,
				TransactionID: "TRX-" + req.OrderID[:8],
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
