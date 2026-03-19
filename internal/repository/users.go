package repository

import (
	"log"
	"military-shop/internal/database"
	"military-shop/internal/models"
	"time"
)

// CreateUser создаёт нового пользователя (с email)
func CreateUser(email, passwordHash, name string) (*models.User, error) {
	var user models.User
	query := `
		INSERT INTO users (email, password_hash, name, phone_verified)
		VALUES ($1, $2, $3, FALSE)
		RETURNING id, email, password_hash, name, COALESCE(phone, '') as phone, phone_verified, is_admin, created_at
	`
	err := database.DB.Get(&user, query, email, passwordHash, name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUserWithPhone создаёт пользователя с подтверждённым телефоном
func CreateUserWithPhone(phone, passwordHash, name string) (*models.User, error) {
	var user models.User
	query := `
		INSERT INTO users (phone, password_hash, name, phone_verified)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id, email, password_hash, name, phone, phone_verified, is_admin, created_at
	`
	err := database.DB.Get(&user, query, phone, passwordHash, name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail возвращает пользователя по email
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, name, COALESCE(phone, '') as phone, phone_verified, is_admin, created_at FROM users WHERE email = $1`
	err := database.DB.Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByPhone возвращает пользователя по телефону
func GetUserByPhone(phone string) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, name, phone, phone_verified, is_admin, created_at FROM users WHERE phone = $1`
	err := database.DB.Get(&user, query, phone)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID возвращает пользователя по ID
func GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, email, password_hash, name, COALESCE(phone, '') as phone, phone_verified, is_admin, created_at FROM users WHERE id = $1`
	err := database.DB.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserPhone обновляет телефон пользователя
func UpdateUserPhone(userID int, phone string, verified bool) error {
	query := `UPDATE users SET phone = $1, phone_verified = $2 WHERE id = $3`
	_, err := database.DB.Exec(query, phone, verified, userID)
	return err
}

// UpdateUserPassword обновляет пароль пользователя
func UpdateUserPassword(userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := database.DB.Exec(query, passwordHash, userID)
	return err
}

// CreateVerificationCode создаёт код верификации (для телефона или email)
func CreateVerificationCode(phone, email, code string, expiresAt time.Time) error {
	// Удаляем старые неиспользованные коды
	if phone != "" {
		_, _ = database.DB.Exec(`DELETE FROM verification_codes WHERE phone = $1 AND used = FALSE`, phone)
	} else if email != "" {
		_, _ = database.DB.Exec(`DELETE FROM verification_codes WHERE email = $1 AND used = FALSE`, email)
	}

	query := `INSERT INTO verification_codes (phone, email, code, expires_at) VALUES ($1, $2, $3, $4)`

	// Обрабатываем пустые строки как NULL для SQL
	var p, e *string
	if phone != "" {
		p = &phone
	}
	if email != "" {
		e = &email
	}

	_, err := database.DB.Exec(query, p, e, code, expiresAt)
	return err
}

// GetValidVerificationCode проверяет и возвращает действующий код
func GetValidVerificationCode(phone, email, code string) (*models.VerificationCode, error) {
	var vc models.VerificationCode

	// Логируем для отладки
	log.Printf("Поиск кода: phone=%s, email=%s, code=%s", phone, email, code)

	var query string
	var err error

	if phone != "" {
		query = `
			SELECT id, phone, email, code, expires_at, used, created_at 
			FROM verification_codes 
			WHERE phone = $1 AND code = $2 AND used = FALSE
		`
		err = database.DB.Get(&vc, query, phone, code)
	} else {
		query = `
			SELECT id, phone, email, code, expires_at, used, created_at 
			FROM verification_codes 
			WHERE email = $1 AND code = $2 AND used = FALSE
		`
		err = database.DB.Get(&vc, query, email, code)
	}

	if err != nil {
		log.Printf("Код не найден: %v", err)
		return nil, err
	}

	// Проверяем срок действия вручную
	if time.Now().After(vc.ExpiresAt) {
		log.Printf("Код просрочен: expires_at=%v, now=%v", vc.ExpiresAt, time.Now())
		return nil, err
	}

	log.Printf("Код найден и действителен: id=%d", vc.ID)
	return &vc, nil
}

// MarkCodeAsUsed помечает код как использованный
func MarkCodeAsUsed(codeID int) error {
	query := `UPDATE verification_codes SET used = TRUE WHERE id = $1`
	_, err := database.DB.Exec(query, codeID)
	return err
}

// GetLatestVerificationCode возвращает последний активный код для проверки попыток
func GetLatestVerificationCode(phone, email string) (*models.VerificationCode, error) {
	var vc models.VerificationCode
	var query string
	var err error

	if phone != "" {
		query = `
			SELECT id, phone, email, code, attempts, expires_at, used, created_at 
			FROM verification_codes 
			WHERE phone = $1 AND used = FALSE
			ORDER BY created_at DESC LIMIT 1
		`
		err = database.DB.Get(&vc, query, phone)
	} else {
		query = `
			SELECT id, phone, email, code, attempts, expires_at, used, created_at 
			FROM verification_codes 
			WHERE email = $1 AND used = FALSE
			ORDER BY created_at DESC LIMIT 1
		`
		err = database.DB.Get(&vc, query, email)
	}

	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// IncrementVerificationAttempts увеличивает счетчик попыток
func IncrementVerificationAttempts(codeID int) error {
	_, err := database.DB.Exec(`UPDATE verification_codes SET attempts = attempts + 1 WHERE id = $1`, codeID)
	return err
}
