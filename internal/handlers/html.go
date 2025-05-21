package handlers

import (
	"html/template"
	"net/http"
)

func ShowPaymentPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("pay").Parse(`
		<!DOCTYPE html>
		<html lang="ru">
		<head>
			<meta charset="UTF-8">
			<title>Оплата</title>
		</head>
		<body>
			<h2>Купить курс — 299 ₽</h2>
			<form action="/pay" method="GET">
				<input type="hidden" name="user_id" value="123456">
				<input type="hidden" name="amount" value="299.00">
				<button type="submit">Перейти к оплате</button>
			</form>
		</body>
		</html>
	`))
	tmpl.Execute(w, nil)
}
