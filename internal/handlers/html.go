package handlers

import (
	"html/template"
	"net/http"
)

func ShowPaymentPage(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	amount := r.URL.Query().Get("amount")
	if userID == "" {
		userID = "0"
	}
	if amount == "" {
		amount = "299" // значение по умолчанию
	}

	tmpl := template.Must(template.New("pay").Parse(`
		<!DOCTYPE html>
		<html lang="ru">
		<head>
			<meta charset="UTF-8">
			<title>Оплата</title>
		</head>
		<body>
			<h2>Купить курс — {{.Amount}} ₽</h2>
			<form action="/pay" method="GET">
				<input type="hidden" name="user_id" value="{{.UserID}}">
				<input type="hidden" name="amount" value="{{.Amount}}">
				<button type="submit">Перейти к оплате</button>
			</form>
		</body>
		</html>
	`))

	data := struct {
		UserID string
		Amount string
	}{
		UserID: userID,
		Amount: amount,
	}

	tmpl.Execute(w, data)
}
