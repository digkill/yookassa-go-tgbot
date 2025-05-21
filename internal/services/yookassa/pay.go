package yookassa

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
)

type CreatePaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	Capture  bool              `json:"capture"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type CreatePaymentResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

func CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	amount := r.URL.Query().Get("amount")
	if userID == "" || amount == "" {
		http.Error(w, "Missing user_id or amount", http.StatusBadRequest)
		return
	}

	shopID := os.Getenv("YOO_SHOP_ID")
	secretKey := os.Getenv("YOO_SECRET_KEY")

	if shopID == "" || secretKey == "" {
		http.Error(w, "Missing YOO_KASSA credentials", http.StatusInternalServerError)
		return
	}

	// Подготовка запроса
	var reqBody CreatePaymentRequest
	reqBody.Amount.Value = amount
	reqBody.Amount.Currency = "RUB"
	reqBody.Confirmation.Type = "redirect"
	reqBody.Confirmation.ReturnURL = "https://yourdomain.ru/success"
	reqBody.Capture = true
	reqBody.Metadata = map[string]string{"user_id": userID}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Создание HTTP-запроса
	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Авторизация через Basic Auth
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", shopID, secretKey)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", uuid.New().String())

	// Выполнение запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to call YooKassa", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Разбор ответа
	var respData CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		http.Error(w, "Invalid response from YooKassa", http.StatusInternalServerError)
		return
	}

	// Редирект на страницу оплаты YooKassa
	if respData.Confirmation.ConfirmationURL != "" {
		http.Redirect(w, r, respData.Confirmation.ConfirmationURL, http.StatusSeeOther)
		return
	}

	http.Error(w, "Missing confirmation_url", http.StatusInternalServerError)
}
