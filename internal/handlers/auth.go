package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"military-shop/internal/middleware"
	"military-shop/internal/models"
	"military-shop/internal/repository"
	"military-shop/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)



// LoginRequest структура запроса авторизации
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// EmailRegisterRequest запрос на начало регистрации через email
type EmailRegisterRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// VerifyEmailRequest запрос на подтверждение email
type VerifyEmailRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// ForgotPasswordRequest запрос на восстановление пароля
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest запрос на сброс пароля
type ResetPasswordRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Code            string `json:"code" binding:"required,len=6"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// ... existing code ...

// Вспомогательная функция: Проверка сложности пароля

func getUserIDFromHeader(c *gin.Context) int {
	// Сначала пробуем из контекста (если сработала middleware)
	if uid, exists := c.Get("user_id"); exists {
		return uid.(int)
	}

	// Иначе проверяем заголовок вручную
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0
	}
	tokenString := parts[1]

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(middleware.GetJWTSecret()), nil
	})

	if err == nil && token.Valid {
		return claims.UserID
	}
	return 0
}

// RegisterEmail начинает процесс регистрации по почте
func RegisterEmail(c *gin.Context) {
	var req EmailRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите корректный email"})
		return
	}

	existingUser, _ := repository.GetUserByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email уже существует"})
		return
	}

	// Генерируем код
	code := services.GenerateCode()
	expiresAt := time.Now().Add(5 * time.Minute)

	// Используем пустой телефон для email-регистрации
	if err := repository.CreateVerificationCode("", req.Email, code, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании кода"})
		return
	}

	emailProvider := services.GetEmailProvider()
	subject := "Подтверждение регистрации"
	message := fmt.Sprintf("Ваш код подтверждения: %s", code)

	if err := emailProvider.SendEmail(req.Email, subject, message); err != nil {
		log.Printf("Failed to send email to %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки Email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Код отправлен на почту",
		"expires_in": 300,
	})
}

// VerifyEmail подтверждает код и создает пользователя
func VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	vc, err := repository.GetLatestVerificationCode("", req.Email)
	if err != nil || vc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код не найден или истёк"})
		return
	}

	// 1. Проверяем количество попыток
	if vc.Attempts >= 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много попыток. Запросите новый код."})
		return
	}

	// 2. Проверяем срок действия
	if time.Now().After(vc.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код просрочен"})
		return
	}

	// 3. Проверяем совпадение кода
	if vc.Code != req.Code {
		_ = repository.IncrementVerificationAttempts(vc.ID)
		remaining := 2 - vc.Attempts
		if remaining < 0 {
			remaining = 0
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код", "attempts_left": remaining})
		return
	}

	// Помечаем код как использованный
	_ = repository.MarkCodeAsUsed(vc.ID)

	existingUser, _ := repository.GetUserByEmail(req.Email)
	if existingUser != nil {
		// Пользователь уже есть, просто логиним
		token, err := generateToken(existingUser.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хеширования пароля"})
		return
	}

	// Создаём пользователя (без телефона, phone_verified=false)
	// Используем уже существующую функцию CreateUser, которая принимает email
	user, err := repository.CreateUser(req.Email, string(hashedPassword), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя"})
		return
	}

	// Генерируем токен
	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Login выполняет вход пользователя
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	var user *models.User
	var err error

	// Ищем пользователя по email
	user, err = repository.GetUserByEmail(req.Email)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль"})
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// GetMe возвращает данные текущего профиля
func GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	user, err := repository.GetUserByID(userID.(int))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ForgotPassword инициирует восстановление пароля
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	var user *models.User
	var err error

	user, err = repository.GetUserByEmail(req.Email)

	if err != nil || user == nil {
		// Логируем для отладки (чтобы понять, почему не отправляется код)
		log.Printf("[ForgotPassword] Пользователь с email %s не найден", req.Email)

		// Не сообщаем, что пользователя нет (защита от перебора аккаунтов)
		c.JSON(http.StatusOK, gin.H{"message": "Если аккаунт существует, код будет отправлен"})
		return
	}

	// Генерируем код
	code := services.GenerateCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	if err := repository.CreateVerificationCode("", req.Email, code, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	emailProvider := services.GetEmailProvider()
	subject := "Восстановление пароля"
	message := fmt.Sprintf("Ваш код для сброса пароля: %s", code)

	if err := emailProvider.SendEmail(req.Email, subject, message); err != nil {
		log.Printf("Failed to send email to %s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки Email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Код отправлен",
		"expires_in": 600,
	})
}

// ResetPassword устанавливает новый пароль по коду
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароли не совпадают"})
		return
	}

	if err := validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль слишком простой: " + err.Error()})
		return
	}

	// Проверяем код (с защитой от перебора)
	var vc *models.VerificationCode
	var err error

	vc, err = repository.GetLatestVerificationCode("", req.Email)
	if err != nil || vc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код не найден или истёк"})
		return
	}

	// 1. Проверяем количество попыток
	if vc.Attempts >= 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много попыток. Запросите новый код."})
		return
	}

	// 2. Проверяем срок действия
	if time.Now().After(vc.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код просрочен"})
		return
	}

	// 3. Проверяем совпадение кода
	if vc.Code != req.Code {
		_ = repository.IncrementVerificationAttempts(vc.ID)
		remaining := 2 - vc.Attempts
		if remaining < 0 {
			remaining = 0
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный код", "attempts_left": remaining})
		return
	}

	// Ищем пользователя
	var user *models.User
	user, err = repository.GetUserByEmail(req.Email)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хэширования"})
		return
	}

	// Обновляем пароль
	if err := repository.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления данных"})
		return
	}

	// Помечаем код как использованный
	_ = repository.MarkCodeAsUsed(vc.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Пароль успешно изменён"})
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("минимум 8 символов")
	}
	// Одна заглавная, одна цифра, один спецсимвол
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	// Спецсимволы (можно расширить)
	specialChars := "!@#$%^&*(),.?\":{}|<>"

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		} else if strings.ContainsRune(specialChars, char) {
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("нужна хотя бы одна заглавная буква")
	}
	if !hasDigit {
		return errors.New("нужна хотя бы одна цифра")
	}
	if !hasSpecial {
		return errors.New("нужен хотя бы один спецсимвол")
	}
	return nil
}



func generateToken(userID int) (string, error) {
	claims := &middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(middleware.GetJWTSecret()))
}
