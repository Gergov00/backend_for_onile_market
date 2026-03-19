package handlers

import (
	"net/http"

	"military-shop/internal/models"
	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetUserOrders возвращает заказы пользователя
func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	orders, err := repository.GetUserOrders(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заказов"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrderByID возвращает детали заказа по ID
func GetOrderByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	orderID := c.Param("id")

	order, err := repository.GetOrderByID(userID.(int), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// CreateOrder создает новый заказ
func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}
	var input models.OrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}

	uid := userID.(int)

	order := models.Order{
		UserID:        &uid,
		CustomerName:  input.CustomerName,
		CustomerPhone: input.CustomerPhone,
		Messenger:     input.Messenger,
		Comment:       input.Comment,
		Address:       input.Address,
		Total:         0, // Итоговая сумма будет рассчитана в репозитории
	}

	items := input.Items
	if len(items) == 0 {
		cartItems, err := repository.GetCartItems(uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения корзины"})
			return
		}
		if len(cartItems) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Корзина пуста"})
			return
		}

		items = make([]models.CartItem, len(cartItems))
		for i, ci := range cartItems {
			items[i] = models.CartItem{
				ID:        ci.ID,
				UserID:    uid,
				ProductID: ci.ProductID,
				Quantity:  ci.Quantity,
			}
		}
	}

	id, total, err := repository.CreateOrder(order, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания заказа: " + err.Error()})
		return
	}

	order.ID = id
	order.Total = total

	// Уведомление удалено (Бот отделен)
	// go service.SendNewOrderNotification(order, items)

	// Очистка корзины при успешном заказе
	if len(input.Items) == 0 { // Если заказ был из корзины
		_ = repository.ClearCart(uid)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Заказ успешно создан"})
}
