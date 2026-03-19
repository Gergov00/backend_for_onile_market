package repository

import (
	"military-shop/internal/database"
)

// CartItemWithProduct элемент корзины с данными товара
type CartItemWithProduct struct {
	ID        int     `json:"id" db:"id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Name      string  `json:"name" db:"name"`
	Price     float64 `json:"price" db:"price"`
	Image     string  `json:"image" db:"image"`
}

// GetCartItems возвращает элементы корзины пользователя
func GetCartItems(userID int) ([]CartItemWithProduct, error) {
	var items []CartItemWithProduct
	query := `
		SELECT 
			ci.id, ci.product_id, ci.quantity,
			p.name, p.price,
			COALESCE(p.images[1], '') as image
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
		ORDER BY ci.id
	`
	err := database.DB.Select(&items, query, userID)
	if err != nil {
		return []CartItemWithProduct{}, err
	}
	return items, nil
}

// AddToCart добавляет товар в корзину
func AddToCart(userID, productID, quantity int) error {
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`
	_, err := database.DB.Exec(query, userID, productID, quantity)
	return err
}

// UpdateCartItem обновляет количество товара в корзине
func UpdateCartItem(userID int, itemID string, quantity int) error {
	query := `UPDATE cart_items SET quantity = $1 WHERE id = $2 AND user_id = $3`
	_, err := database.DB.Exec(query, quantity, itemID, userID)
	return err
}

// RemoveFromCart удаляет товар из корзины
func RemoveFromCart(userID int, itemID string) error {
	query := `DELETE FROM cart_items WHERE id = $1 AND user_id = $2`
	_, err := database.DB.Exec(query, itemID, userID)
	return err
}

// ClearCart очищает корзину пользователя
func ClearCart(userID int) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := database.DB.Exec(query, userID)
	return err
}

// GetCartCount возвращает количество товаров в корзине
func GetCartCount(userID int) (int, error) {
	var count int
	query := `SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id = $1`
	err := database.DB.Get(&count, query, userID)
	return count, err
}
