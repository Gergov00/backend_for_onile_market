package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"military-shop/internal/repository"

	"github.com/gin-gonic/gin"
)

// cyrillicToLatin - таблица транслитерации для генерации slug
var cyrillicToLatin = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// transliterate преобразует кириллический текст в URL-совместимый slug
func transliterate(text string) string {
	text = strings.ToLower(text)
	var result strings.Builder
	for _, r := range text {
		if val, ok := cyrillicToLatin[r]; ok {
			result.WriteString(val)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}
	// Убираем двойные дефисы и дефисы по краям
	slug := result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	// Ограничиваем длину
	if len(slug) > 60 {
		slug = slug[:60]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

// GenerateSitemap автоматически генерирует sitemap.xml на основе данных из БД
// GET /api/sitemap
func GenerateSitemap(c *gin.Context) {
	baseURL := "https://russian-armo.org"
	now := time.Now().Format("2006-01-02")

	var urls []string

	// Статические страницы
	staticPages := []struct {
		loc        string
		changefreq string
		priority   string
	}{
		{"/", "daily", "1.0"},
		{"/categories", "weekly", "0.9"},
		{"/gallery", "weekly", "0.8"},
		{"/contacts", "monthly", "0.6"},
		{"/privacy", "monthly", "0.3"},
		{"/offer", "monthly", "0.3"},
	}

	for _, page := range staticPages {
		urls = append(urls, fmt.Sprintf(`  <url>
    <loc>%s%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>%s</changefreq>
    <priority>%s</priority>
  </url>`, baseURL, page.loc, now, page.changefreq, page.priority))
	}

	// Категории
	categories, err := repository.GetAllCategories()
	if err == nil {
		for _, cat := range categories {
			urls = append(urls, fmt.Sprintf(`  <url>
    <loc>%s/category/%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>`, baseURL, cat.Slug, now))
		}
	}

	// Товары
	products, err := repository.GetAllProducts()
	if err == nil {
		for _, p := range products {
			slug := transliterate(p.Name)
			productURL := fmt.Sprintf("%s/product/%s-%d", baseURL, slug, p.ID)
			urls = append(urls, fmt.Sprintf(`  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.7</priority>
  </url>`, productURL, now))
		}
	}

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
%s
</urlset>`, strings.Join(urls, "\n"))

	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(sitemap))
}
