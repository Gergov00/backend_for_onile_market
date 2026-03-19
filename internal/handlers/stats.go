package handlers

import (
	"net/http"

	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetAdminStats возвращает общую статистику для панели управления
func GetAdminStats(c *gin.Context) {
	stats, err := repository.GetSiteStats(7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения статистики посещений"})
		return
	}

	topProducts, err := repository.GetTopViewedProducts(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения популярных товаров"})
		return
	}

	products, err := repository.GetAllProducts()
	totalProducts := len(products)

	totalViews := 0
	for _, p := range products {
		totalViews += p.Views
	}

	c.JSON(http.StatusOK, gin.H{
		"site_stats":     stats,
		"top_products":   topProducts,
		"total_products": totalProducts,
	})
}

// RecordVisit регистрирует визит на сайт
func RecordVisit(c *gin.Context) {
	if err := repository.IncrementSiteVisits(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка записи посещения"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
