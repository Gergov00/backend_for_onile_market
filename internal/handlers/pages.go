package handlers

import (
	"log"
	"net/http"

	"military-shop/internal/database"
	"military-shop/internal/models"

	"github.com/gin-gonic/gin"
)

// GetPages возвращает список всех страниц (только название и slug для меню админки)
func GetPages(c *gin.Context) {
	var pages []models.Page
	query := "SELECT id, slug, title, updated_at FROM pages ORDER BY title"
	err := database.DB.Select(&pages, query)
	if err != nil {
		log.Printf("GetPages error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения страниц"})
		return
	}
	c.JSON(http.StatusOK, pages)
}

// GetPageBySlug возвращает полную страницу по slug
func GetPageBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var page models.Page
	query := "SELECT id, slug, title, content, updated_at FROM pages WHERE slug = $1"
	err := database.DB.Get(&page, query, slug)
	if err != nil {
		log.Printf("GetPageBySlug error: %v, slug: %s", err, slug)
		c.JSON(http.StatusNotFound, gin.H{"error": "Страница не найдена"})
		return
	}
	c.JSON(http.StatusOK, page)
}

// UpdatePage обновляет контент страницы
func UpdatePage(c *gin.Context) {
	slug := c.Param("slug")
	
	var page models.Page
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных: " + err.Error()})
		return
	}

	query := `
		UPDATE pages 
		SET content = $1, title = $2, updated_at = CURRENT_TIMESTAMP
		WHERE slug = $3
	`

	result, err := database.DB.Exec(query, page.Content, page.Title, slug)
	if err != nil {
		log.Printf("UpdatePage error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления страницы"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Страница не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Страница успешно обновлена"})
}
