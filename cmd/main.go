package main

import (
	"gitlab.com/mediarise/yookassa-go/internal/handlers"
	"gitlab.com/mediarise/yookassa-go/internal/services/yookassa"
	"log"
	"net/http"
	"os"
)

func main() {
	// 💳 Обработка вебхука от Юкассы
	http.HandleFunc("/webhook", yookassa.WebhookHandler)

	// 💵 Страница оплаты (GET /)
	http.HandleFunc("/", handlers.ShowPaymentPage)

	// 💸 Обработка перехода к оплате (GET /pay)
	http.HandleFunc("/pay", yookassa.CreatePaymentHandler)

	// 🌐 Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8383"
	}

	log.Println("✅ Server listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
