package handlers

import (
	"net/http"

	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)



// AddToCartRequest запрос на добавление в корзину
type AddToCartRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity"`
}

// UpdateCartRequest запрос на обновление количества
type UpdateCartRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// GetCart возвращает корзину текущего пользователя
func GetCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	items, err := repository.GetCartItems(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения корзины"})
		return
	}

	// Считаем итого
	var total float64
	var count int
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
		count += item.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"count": count,
	})
}

// AddToCart добавляет товар в корзину пользователя
func AddToCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите ID товара"})
		return
	}

	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	err := repository.AddToCart(userID.(int), req.ProductID, quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка добавления в корзину"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар добавлен в корзину"})
}

// UpdateCartItem изменяет количество товара в корзине
func UpdateCartItem(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	itemID := c.Param("id")

	var req UpdateCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите количество"})
		return
	}

	err := repository.UpdateCartItem(userID.(int), itemID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Количество обновлено"})
}

// RemoveFromCart удаляет позицию из корзины
func RemoveFromCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	itemID := c.Param("id")

	err := repository.RemoveFromCart(userID.(int), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар удалён из корзины"})
}

// ClearCart полностью очищает корзину пользователя
func ClearCart(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	err := repository.ClearCart(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка очистки корзины"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Корзина очищена"})
}
