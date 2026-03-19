package repository

import (
	"military-shop/internal/database"
	"military-shop/internal/models"
	"time"
)

// IncrementSiteVisits увеличивает счетчик посещений сайта за сегодня
func IncrementSiteVisits() error {
	date := time.Now().Format("2006-01-02")
	query := `
		INSERT INTO site_stats (date, visits)
		VALUES ($1, 1)
		ON CONFLICT (date)
		DO UPDATE SET visits = site_stats.visits + 1
	`
	_, err := database.DB.Exec(query, date)
	return err
}

// GetSiteStats возвращает статистику посещений за последние N дней
func GetSiteStats(limit int) ([]models.SiteStats, error) {
	var stats []models.SiteStats
	query := `SELECT id, date, visits, created_at FROM site_stats ORDER BY date DESC LIMIT $1`
	err := database.DB.Select(&stats, query, limit)
	return stats, err
}
