package handlers

import (
	"net/http"
	"strconv"

	"military-shop/internal/models"
	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetGalleryItems получить все элементы галереи (публичный)
func GetGalleryItems(c *gin.Context) {
	itemType := c.Query("type") // "photo", "video" или пустая строка

	var items []models.GalleryItem
	var err error

	if itemType != "" {
		items, err = repository.GetGalleryItemsByType(itemType)
	} else {
		items, err = repository.GetGalleryItems()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка загрузки галереи"})
		return
	}

	if items == nil {
		items = []models.GalleryItem{}
	}

	c.JSON(http.StatusOK, items)
}

// CreateGalleryItem создать элемент галереи (админ)
func CreateGalleryItem(c *gin.Context) {
	var item models.GalleryItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные: " + err.Error()})
		return
	}

	if item.Type != "photo" && item.Type != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Тип должен быть 'photo' или 'video'"})
		return
	}

	if item.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL обязателен"})
		return
	}

	if err := repository.CreateGalleryItem(&item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания элемента галереи"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// DeleteGalleryItem удалить элемент галереи (админ)
func DeleteGalleryItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := repository.DeleteGalleryItem(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Элемент удален"})
}
