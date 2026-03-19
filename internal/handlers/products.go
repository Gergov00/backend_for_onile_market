package handlers

import (
	"net/http"
	"strconv"

	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetCategories возвращает список всех категорий
func GetCategories(c *gin.Context) {
	categories, err := repository.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения категорий"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// GetCategoryBySlug возвращает категорию и ее товары по слагу
func GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := repository.GetCategoryBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Категория не найдена"})
		return
	}

	products, err := repository.GetProductsByCategory(category.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения товаров"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
		"products": products,
	})
}

// GetProducts возвращает список всех товаров
func GetProducts(c *gin.Context) {
	products, err := repository.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения товаров"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// GetProductByID возвращает детальную информацию о товаре
func GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID товара"})
		return
	}

	product, err := repository.GetProductByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	go func() {
		_ = repository.IncrementProductViews(id)
	}()

	c.JSON(http.StatusOK, product)
}

// SearchProducts выполняет поиск товаров
func SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пустой поисковый запрос"})
		return
	}

	products, err := repository.SearchProducts(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка поиска товаров"})
		return
	}

	c.JSON(http.StatusOK, products)
}
