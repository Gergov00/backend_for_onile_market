package repository

import (
	"fmt"
	"math/rand"
	"military-shop/internal/database"
	"military-shop/internal/models"
	"time"
)

func CreateOrder(order models.Order, items []models.CartItem) (int, float64, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	var total float64
	type itemData struct {
		ProductID int
		Quantity  int
		Price     float64
	}
	var preparedItems []itemData

	for _, item := range items {
		var price float64
		err := tx.QueryRow("SELECT price FROM products WHERE id = $1", item.ProductID).Scan(&price)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get price for product %d: %w", item.ProductID, err)
		}
		total += price * float64(item.Quantity)
		preparedItems = append(preparedItems, itemData{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		})
	}

	var orderID int

	// Генерация случайного ID от 100000000 до 2000000000
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	orderID = rng.Intn(1900000000) + 100000000

	// Использование явной вставки ID
	query := `
		INSERT INTO orders (id, user_id, customer_name, customer_phone, messenger, comment, total, address, status, taken_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'new', NULL, NOW())
		RETURNING id`

	err = tx.QueryRow(query,
		orderID,
		order.UserID,
		order.CustomerName,
		order.CustomerPhone,
		order.Messenger,
		order.Comment,
		total,
		order.Address,
	).Scan(&orderID)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to insert order: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	for _, item := range preparedItems {
		_, err = stmt.Exec(orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return orderID, total, tx.Commit()
}

// UpdateOrderStatus обновляет статус заказа
func UpdateOrderStatus(orderID int, status string) error {
	_, err := database.DB.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	return err
}

// SetOrderTaken помечает заказ как обрабатываемый, если он сейчас в статусе "new"
func SetOrderTaken(orderID int, adminName string) error {
	res, err := database.DB.Exec("UPDATE orders SET status = 'processing', taken_by = $1 WHERE id = $2 AND status = 'new'", adminName, orderID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("order not found or already taken")
	}
	return nil
}

// GetOrder возвращает один заказ по ID
func GetOrder(orderID int) (models.Order, error) {
	var o models.Order
	err := database.DB.Get(&o, "SELECT * FROM orders WHERE id = $1", orderID)
	return o, err
}

// OrderFilter определяет критерии для получения заказов
type OrderFilter struct {
	Statuses  []string
	TakenBy   *string // nil = любой, "" = без назначенного админа
	Limit     int
	Offset    int
	SortOrder string // "ASC" or "DESC"
}

// GetOrders возвращает заказы, соответствующие фильтру
func GetOrders(filter OrderFilter) ([]models.Order, error) {
	query := "SELECT * FROM orders WHERE 1=1"
	args := []interface{}{}
	argID := 1

	// Сборка запроса с фильтрацией по статусам
	if len(filter.Statuses) > 0 {
		statusPlaceholders := ""
		for i, status := range filter.Statuses {
			if i > 0 {
				statusPlaceholders += ", "
			}
			statusPlaceholders += fmt.Sprintf("$%d", argID)
			args = append(args, status)
			argID++
		}
		query += fmt.Sprintf(" AND status IN (%s)", statusPlaceholders)
	}

	if filter.TakenBy != nil {
		query += fmt.Sprintf(" AND taken_by = $%d", argID)
		args = append(args, *filter.TakenBy)
		argID++
	}

	if filter.SortOrder == "ASC" {
		query += " ORDER BY created_at ASC"
	} else {
		query += " ORDER BY created_at DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	var orders []models.Order
	err := database.DB.Select(&orders, query, args...)
	return orders, err
}

// OrderWithItems заказ с товарами
type OrderWithItems struct {
	ID        int         `json:"id" db:"id"`
	Total     float64     `json:"total" db:"total"`
	Status    string      `json:"status" db:"status"`
	Address   string      `json:"address" db:"address"`
	CreatedAt string      `json:"created_at" db:"created_at"`
	Items     []OrderItem `json:"items"`
}

// OrderItem позиция заказа
type OrderItem struct {
	ID        int     `json:"id" db:"id"`
	ProductID int     `json:"product_id" db:"product_id"`
	Name      string  `json:"name" db:"name"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
	Image     string  `json:"image" db:"image"`
}

// GetUserOrders возвращает заказы пользователя
func GetUserOrders(userID int) ([]OrderWithItems, error) {
	var orders []OrderWithItems
	query := `
		SELECT id, total, status, COALESCE(address, '') as address, 
		       TO_CHAR(created_at, 'DD.MM.YYYY HH24:MI') as created_at
		FROM orders 
		WHERE user_id = $1 
		ORDER BY created_at DESC
	`
	err := database.DB.Select(&orders, query, userID)
	if err != nil {
		return []OrderWithItems{}, err
	}

	// Получаем позиции для каждого заказа
	for i := range orders {
		items, _ := GetOrderItems(orders[i].ID)
		orders[i].Items = items
	}

	return orders, nil
}

// GetOrderItems возвращает позиции заказа
func GetOrderItems(orderID int) ([]OrderItem, error) {
	var items []OrderItem
	query := `
		SELECT oi.id, oi.product_id, COALESCE(p.name, 'Товар удалён') as name, 
		       oi.quantity, oi.price, COALESCE(p.images[1], '') as image
		FROM order_items oi
		LEFT JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`
	err := database.DB.Select(&items, query, orderID)
	if err != nil {
		return []OrderItem{}, err
	}
	return items, nil
}

// GetOrderByID возвращает заказ по ID (для пользователя)
func GetOrderByID(userID int, orderID string) (*OrderWithItems, error) {
	var order OrderWithItems
	query := `
		SELECT id, total, status, COALESCE(address, '') as address,
		       TO_CHAR(created_at, 'DD.MM.YYYY HH24:MI') as created_at
		FROM orders 
		WHERE id = $1 AND user_id = $2
	`
	err := database.DB.Get(&order, query, orderID, userID)
	if err != nil {
		return nil, err
	}

	order.Items, _ = GetOrderItems(order.ID)
	return &order, nil
}
