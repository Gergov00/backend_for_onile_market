package models

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// Page представляет статичную страницу (например, оферта)
type Page struct {
	ID        int       `json:"id" db:"id"`
	Slug      string    `json:"slug" db:"slug"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Category представляет категорию товаров
type Category struct {
	ID           int    `json:"id" db:"id"`
	Name         string `json:"name" db:"name"`
	Slug         string `json:"slug" db:"slug"`
	Description  string `json:"description" db:"description"`
	Image        string `json:"image" db:"image"`
	Count        int    `json:"count" db:"count"`
	ShowInHeader bool   `json:"show_in_header" db:"show_in_header"`
}

// Product представляет товар
type Product struct {
	ID          int             `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	CategoryID  int             `json:"category_id" db:"category_id"`
	Category    string          `json:"category" db:"category_name"`
	Price       float64         `json:"price" db:"price"`
	OldPrice    sql.NullFloat64 `json:"-" db:"old_price"`
	OldPriceVal *float64        `json:"oldPrice"`
	Badge       sql.NullString  `json:"-" db:"badge"`
	BadgeVal    *string         `json:"badge"`
	IsNew       bool            `json:"isNew" db:"is_new"`
	Images      pq.StringArray  `json:"images" db:"images"`
	Documents   pq.StringArray  `json:"documents" db:"documents"`
	Description string          `json:"description" db:"description"`
	Specs       json.RawMessage `json:"specs" db:"specs"`
	Views       int             `json:"views" db:"views"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// SiteStats статистика сайта
type SiteStats struct {
	ID        int       `db:"id"`
	Date      time.Time `db:"date"`
	Visits    int       `db:"visits"`
	CreatedAt time.Time `db:"created_at"`
}

// ProcessNullFields преобразует NULL поля в указатели
func (p *Product) ProcessNullFields() {
	if p.OldPrice.Valid {
		p.OldPriceVal = &p.OldPrice.Float64
	}
	if p.Badge.Valid {
		p.BadgeVal = &p.Badge.String
	}
}

// User представляет пользователя
type User struct {
	ID            int            `json:"id" db:"id"`
	Email         sql.NullString `json:"email,omitempty" db:"email"`
	PasswordHash  string         `json:"-" db:"password_hash"`
	Name          string         `json:"name" db:"name"`
	Phone         string         `json:"phone" db:"phone"`
	PhoneVerified bool           `json:"phone_verified" db:"phone_verified"`
	IsAdmin       bool           `json:"is_admin" db:"is_admin"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}

// VerificationCode представляет код подтверждения телефона или почты
type VerificationCode struct {
	ID        int            `db:"id"`
	Phone     sql.NullString `db:"phone"`
	Email     sql.NullString `db:"email"`
	Code      string         `db:"code"`
	Attempts  int            `db:"attempts"`
	ExpiresAt time.Time      `db:"expires_at"`
	Used      bool           `db:"used"`
	CreatedAt time.Time      `db:"created_at"`
}

// Order представляет заказ
type Order struct {
	ID            int            `json:"id" db:"id"`
	UserID        *int           `json:"user_id" db:"user_id"` // Может быть NULL для гостевых заказов
	CustomerName  string         `json:"customer_name" db:"customer_name"`
	CustomerPhone string         `json:"customer_phone" db:"customer_phone"`
	Messenger     string         `json:"messenger" db:"messenger"`
	Comment       string         `json:"comment" db:"comment"`
	Total         float64        `json:"total" db:"total"`
	Status        string         `json:"status" db:"status"`
	Address       string         `json:"address" db:"address"`
	TakenBy       sql.NullString `json:"taken_by" db:"taken_by"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}

// OrderInput структура для создания заказа
type OrderInput struct {
	CustomerName  string     `json:"customer_name" binding:"required"`
	CustomerPhone string     `json:"customer_phone" binding:"required"`
	Messenger     string     `json:"messenger" binding:"required"`
	Comment       string     `json:"comment"`
	Address       string     `json:"address"`
	Items         []CartItem `json:"items"` // Опционально: если передано напрямую, иначе из корзины
}

// OrderItem представляет позицию заказа
type OrderItem struct {
	ID        int     `json:"id" db:"id"`
	OrderID   int     `json:"order_id" db:"order_id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
}

// CartItem представляет товар в корзине
type CartItem struct {
	ID        int `json:"id" db:"id"`
	UserID    int `json:"user_id" db:"user_id"`
	ProductID int `json:"product_id" db:"product_id"`
	Quantity  int `json:"quantity" db:"quantity"`
}

// GalleryItem представляет элемент галереи (фото или видео)
type GalleryItem struct {
	ID          int       `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"`
	URL         string    `json:"url" db:"url"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
