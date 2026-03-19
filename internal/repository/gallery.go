package repository

import (
	"military-shop/internal/database"
	"military-shop/internal/models"
)

// GetGalleryItems возвращает все элементы галереи
func GetGalleryItems() ([]models.GalleryItem, error) {
	var items []models.GalleryItem
	query := `SELECT id, type, url, COALESCE(title, '') as title, COALESCE(description, '') as description, sort_order, created_at 
			  FROM gallery_items ORDER BY sort_order ASC, created_at DESC`
	err := database.DB.Select(&items, query)
	return items, err
}

// GetGalleryItemsByType возвращает элементы по типу (photo/video)
func GetGalleryItemsByType(itemType string) ([]models.GalleryItem, error) {
	var items []models.GalleryItem
	query := `SELECT id, type, url, COALESCE(title, '') as title, COALESCE(description, '') as description, sort_order, created_at 
			  FROM gallery_items WHERE type = $1 ORDER BY sort_order ASC, created_at DESC`
	err := database.DB.Select(&items, query, itemType)
	return items, err
}

// CreateGalleryItem создает новый элемент галереи
func CreateGalleryItem(item *models.GalleryItem) error {
	query := `INSERT INTO gallery_items (type, url, title, description, sort_order, created_at)
			  VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id`
	return database.DB.QueryRow(query, item.Type, item.URL, item.Title, item.Description, item.SortOrder).Scan(&item.ID)
}

// DeleteGalleryItem удаляет элемент галереи
func DeleteGalleryItem(id int) error {
	query := `DELETE FROM gallery_items WHERE id = $1`
	_, err := database.DB.Exec(query, id)
	return err
}
