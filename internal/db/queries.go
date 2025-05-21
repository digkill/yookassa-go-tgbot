package db

import (
	"log"
	"time"
)

func CreateUserIfNotExists(telegramID int64) (int, error) {
	var id int
	err := DB.QueryRow(`
        INSERT INTO users (telegram_id)
        VALUES ($1)
        ON CONFLICT (telegram_id) DO UPDATE SET telegram_id = EXCLUDED.telegram_id
        RETURNING id
    `, telegramID).Scan(&id)
	return id, err
}

func CreatePayment(userID int, yookassaID string, amount float64) error {
	_, err := DB.Exec(`
        INSERT INTO payments (user_id, yookassa_payment_id, amount, status)
        VALUES ($1, $2, $3, 'pending')
    `, userID, yookassaID, amount)
	return err
}

func UpdatePaymentStatus(yookassaID string, status string) error {
	_, err := DB.Exec(`
        UPDATE payments
        SET status = $1, updated_at = $2
        WHERE yookassa_payment_id = $3
    `, status, time.Now(), yookassaID)
	return err
}

func MarkUserAsPaid(userID int64) error {
	_, err := DB.Exec(`
        UPDATE users
        SET subscription_ends_at = NOW() + interval '30 days'
        WHERE chat_id = $1
    `, userID)

	if err != nil {
		log.Printf("❌ Ошибка при установке подписки для пользователя %d: %v", userID, err)
		return err
	}

	log.Printf("✅ Пользователь %d получил подписку на 30 дней", userID)
	return nil
}

func GetPaidUsers() ([]int64, error) {
	rows, err := DB.Query(`SELECT telegram_id FROM users WHERE is_paid = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GrantAccessAfterPayment симулирует выдачу доступа к курсу
func GrantAccessAfterPayment(telegramID int64) error {
	// Здесь может быть вызов к боту, email-рассылка, запись в другую таблицу и т.д.
	log.Printf("🎁 Пользователю %d предоставлен доступ к курсу", telegramID)
	return nil
}

func HasActiveSubscription(userID int64) (bool, error) {
	var endsAt *time.Time
	err := DB.QueryRow(`SELECT subscription_ends_at FROM users WHERE chat_id = $1`, userID).Scan(&endsAt)

	if err != nil {
		log.Printf("Ошибка при проверке подписки у пользователя %d: %v", userID, err)
		return false, err
	}

	if endsAt == nil {
		return false, nil
	}

	return endsAt.After(time.Now()), nil
}
