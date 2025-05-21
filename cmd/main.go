package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gitlab.com/mediarise/yookassa-go/internal/db"
	"gitlab.com/mediarise/yookassa-go/internal/handlers"
	"gitlab.com/mediarise/yookassa-go/internal/services/yookassa"
	"log"
	"net/http"
	"os"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	if err := godotenv.Load(); err != nil {
		logrus.Warnf("load env failed: %v", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	db.Init(dsn)

	// Обработка вебхука от Юкассы
	http.HandleFunc("/webhook", yookassa.WebhookHandler)

	// Страница оплаты (GET /)
	http.HandleFunc("/", handlers.ShowPaymentPage)

	// Обработка перехода к оплате (GET /pay)
	http.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreatePaymentHandler(w, r, &db)
	})

	// Обработка успешного платежа
	http.HandleFunc("/success", yookassa.SuccessHandler)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8383"
	}

	log.Println("✅ Server listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
