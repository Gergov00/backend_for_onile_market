package handlers

import (
	"net/http"
	"strconv"

	"military-shop/internal/database"
	"military-shop/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// CreateProduct создает новый товар
func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var oldPrice *float64
	if product.OldPriceVal != nil {
		oldPrice = product.OldPriceVal
	}
	var badge *string
	if product.BadgeVal != nil && *product.BadgeVal != "" {
		badge = product.BadgeVal
	}

	query := `
		INSERT INTO products (name, category_id, price, old_price, description, images, documents, specs, badge, is_new, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id
	`

	// Преобразуем images, documents и specs для pg
	imagesArray := pq.Array(product.Images)
	documentsArray := pq.Array(product.Documents)
	// specs храним как jsonb

	err := database.DB.QueryRow(query,
		product.Name,
		product.CategoryID,
		product.Price,
		oldPrice,
		product.Description,
		imagesArray,
		documentsArray,
		product.Specs,
		badge,
		product.IsNew,
	).Scan(&product.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания товара: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct обновляет товар
func UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var oldPrice *float64
	if product.OldPriceVal != nil {
		oldPrice = product.OldPriceVal
	}
	var badge *string
	if product.BadgeVal != nil && *product.BadgeVal != "" {
		badge = product.BadgeVal
	}

	query := `
		UPDATE products 
		SET name=$1, category_id=$2, price=$3, old_price=$4, description=$5, images=$6, documents=$7, specs=$8, badge=$9, is_new=$10
		WHERE id=$11
	`

	imagesArray := pq.Array(product.Images)
	documentsArray := pq.Array(product.Documents)

	result, err := database.DB.Exec(query,
		product.Name,
		product.CategoryID,
		product.Price,
		oldPrice,
		product.Description,
		imagesArray,
		documentsArray,
		product.Specs,
		badge,
		product.IsNew,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления товара"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар обновлен"})
}

// DeleteProduct удаляет товар
func DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	query := `DELETE FROM products WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления товара"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Товар удален"})
}

// CreateCategory создает новую категорию
func CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO categories (name, slug, description, image, show_in_header)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := database.DB.QueryRow(query,
		category.Name,
		category.Slug,
		category.Description,
		category.Image,
		category.ShowInHeader,
	).Scan(&category.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания категории: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory обновляет категорию
func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE categories 
		SET name=$1, slug=$2, description=$3, image=$4, show_in_header=$5
		WHERE id=$6
	`

	result, err := database.DB.Exec(query,
		category.Name,
		category.Slug,
		category.Description,
		category.Image,
		category.ShowInHeader,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления категории: " + err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Категория не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Категория обновлена"})
}

// DeleteCategory удаляет категорию
func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	query := `DELETE FROM categories WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления категории, возможно к ней привязаны товары: " + err.Error()})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Категория не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Категория удалена"})
}
