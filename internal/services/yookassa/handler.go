package yookassa

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type PaymentNotification struct {
	Event  string `json:"event"`
	Object struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Metadata map[string]string `json:"metadata"`
	} `json:"object"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var notif PaymentNotification
	if err := json.Unmarshal(body, &notif); err != nil {
		log.Printf("❌ Bad JSON: %s", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("📥 Received payment: %s %s [%s]",
		notif.Object.Amount.Value,
		notif.Object.Amount.Currency,
		notif.Object.Status,
	)

	// Пример логики: если платёж успешен
	if notif.Object.Status == "succeeded" {
		userID := notif.Object.Metadata["user_id"]
		log.Printf("🎉 Payment success for user %s", userID)
		// Можно: уведомить Telegram, обновить БД и т.д.
	}

	w.WriteHeader(http.StatusOK)
}
