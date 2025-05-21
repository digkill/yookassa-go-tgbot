package yookassa

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/mediarise/yookassa-go/internal/db"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
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
	Metadata map[string]string `json:"metadata"`
}

type CreatePaymentResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

func CreatePaymentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	userID := r.URL.Query().Get("user_id")
	amount := r.URL.Query().Get("amount")
	if userID == "" || amount == "" {
		http.Error(w, "Missing user_id or amount", http.StatusBadRequest)
		return
	}

	shopID := os.Getenv("YOO_SHOP_ID")
	secretKey := os.Getenv("YOO_SECRET_KEY")
	successURL := os.Getenv("YOO_SUCCESS_URL")

	if shopID == "" || secretKey == "" || successURL == "" {
		http.Error(w, "Missing Yookassa env variables", http.StatusInternalServerError)
		return
	}

	orderID := uuid.New().String()

	// BEGIN TRANSACTION
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	// Сохраняем PENDING-платеж
	_, err = tx.Exec(`
		INSERT INTO payments (user_id, order_id, amount, status)
		VALUES ($1, $2, $3, 'pending')
	`, userID, orderID, amount)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to insert payment record", http.StatusInternalServerError)
		return
	}

	// Подготовка запроса в YooKassa
	var reqBody CreatePaymentRequest
	reqBody.Amount.Value = amount
	reqBody.Amount.Currency = "RUB"
	reqBody.Confirmation.Type = "redirect"
	reqBody.Confirmation.ReturnURL = fmt.Sprintf("%s?user_id=%s", successURL, userID)
	reqBody.Capture = true
	reqBody.Metadata = map[string]string{
		"user_id":  userID,
		"order_id": orderID,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}

	// Отправка запроса в YooKassa
	req, err := http.NewRequest("POST", "https://api.yookassa.ru/v3/payments", bytes.NewBuffer(payload))
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to create YooKassa request", http.StatusInternalServerError)
		return
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", shopID, secretKey)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", uuid.New().String())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to contact YooKassa", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var respData CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		tx.Rollback()
		http.Error(w, "Invalid response from YooKassa", http.StatusInternalServerError)
		return
	}

	// Обновляем запись — записываем payment_id
	_, err = tx.Exec(`
		UPDATE payments
		SET yookassa_payment_id = $1, updated_at = now()
		WHERE order_id = $2
	`, respData.ID, orderID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Failed to update payment record", http.StatusInternalServerError)
		return
	}

	// COMMIT
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit payment transaction", http.StatusInternalServerError)
		return
	}

	// Редирект на оплату
	if respData.Confirmation.ConfirmationURL != "" {
		http.Redirect(w, r, respData.Confirmation.ConfirmationURL, http.StatusSeeOther)
		return
	}

	http.Error(w, "Missing confirmation_url", http.StatusInternalServerError)
}

func SuccessHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID64, _ := strconv.ParseInt(userID, 10, 64)
	isActive, err := db.HasActiveSubscription(userID64)
	if err != nil {
		http.Error(w, "🚫 Ошибка при проверке подписки", http.StatusBadRequest)
		return
	}

	if isActive == true {
		http.Error(w, "Подписка уже активна", http.StatusInternalServerError)
		return
	}

	if err := db.MarkUserAsPaid(userID64); err != nil {
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("success").Parse(`
<!DOCTYPE html>
<html lang="ru">
<head><meta charset="UTF-8"><title>Оплата успешна</title></head>
<body>
    <h2>Спасибо за оплату!</h2>
    <p>Доступ к курсу открыт для пользователя с ID: {{.}}</p>
</body>
</html>`))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, userID)
}
