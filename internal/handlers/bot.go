package handlers

import (
	"net/http"
	"strconv"

	"military-shop/internal/models"
	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetBotOrders возвращает список заказов для бота
func GetBotOrders(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.Query("limit")
	limit := 20 // По умолчанию для бота

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	var statuses []string
	if status != "" {
		statuses = []string{status}
	}

	filter := repository.OrderFilter{
		Statuses:  statuses,
		Limit:     limit,
		SortOrder: "DESC",
	}

	takenBy := c.Query("taken_by")
	if takenBy != "" {
		filter.TakenBy = &takenBy
	}

	orders, err := repository.GetOrders(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения заказов"})
		return
	}

	// Дополняем заказы товарами
	type OrderWithItems struct {
		models.Order
		Items []repository.OrderItem `json:"items"`
	}

	var result []OrderWithItems
	for _, order := range orders {
		items, _ := repository.GetOrderItems(order.ID)
		result = append(result, OrderWithItems{
			Order: order,
			Items: items,
		})
	}

	c.JSON(http.StatusOK, result)
}

// GetBotOrder возвращает детали заказа по ID
func GetBotOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	order, err := repository.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	items, err := repository.GetOrderItems(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения товаров"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
		"items": items,
	})
}

// BotTakeOrder назначает администратора на заказ
func BotTakeOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var req struct {
		AdminName string `json:"admin_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err := repository.SetOrderTaken(orderID, req.AdminName); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// BotUpdateStatus изменяет текущий статус заказа
func BotUpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	if err := repository.UpdateOrderStatus(orderID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления статуса"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
