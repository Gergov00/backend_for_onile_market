package repository

import (
	"fmt"
	"strings"

	"military-shop/internal/database"
	"military-shop/internal/models"
)

// GetAllCategories возвращает все доступные категории
func GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.image, c.show_in_header, COUNT(p.id) as count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id
		GROUP BY c.id, c.name, c.slug, c.description, c.image, c.show_in_header
		ORDER BY c.id
	`
	err := database.DB.Select(&categories, query)
	return categories, err
}

// GetCategoryBySlug ищет категорию по ее слагу
func GetCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.image, c.show_in_header, COUNT(p.id) as count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id
		WHERE c.slug = $1
		GROUP BY c.id, c.name, c.slug, c.description, c.image, c.show_in_header
	`
	err := database.DB.Get(&category, query, slug)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetAllProducts возвращает список всех товаров
func GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT 
			p.id, p.name, p.category_id, c.name as category_name,
			p.price, p.old_price, p.badge, p.is_new, 
			p.images, p.documents, p.description, p.specs, p.created_at, p.views
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.id
	`
	err := database.DB.Select(&products, query)
	if err != nil {
		return nil, err
	}

	// Обработка NULL полей
	for i := range products {
		products[i].ProcessNullFields()
	}

	return products, nil
}

// GetProductByID ищет товар по числовому ID
func GetProductByID(id int) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT 
			p.id, p.name, p.category_id, c.name as category_name,
			p.price, p.old_price, p.badge, p.is_new, 
			p.images, p.documents, p.description, p.specs, p.created_at, p.views
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`
	err := database.DB.Get(&product, query, id)
	if err != nil {
		return nil, err
	}

	product.ProcessNullFields()
	return &product, nil
}

// GetProductsByCategory возвращает товары, привязанные к категории
func GetProductsByCategory(categoryID int) ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT 
			p.id, p.name, p.category_id, c.name as category_name,
			p.price, p.old_price, p.badge, p.is_new, 
			p.images, p.documents, p.description, p.specs, p.created_at, p.views
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1
		ORDER BY p.id
	`
	err := database.DB.Select(&products, query, categoryID)
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].ProcessNullFields()
	}

	return products, nil
}

// getSearchVariations генерирует варианты для поиска, включая синонимы
func getSearchVariations(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	variationsMap := map[string]bool{
		query:               true,
		switchLayout(query): true,
	}

	// Расширенный словарь синонимов и частых опечаток
	synonyms := map[string][]string{
		"мавик":         {"mavic", "dji mavic", "квадрокоптер", "дрон"},
		"mavic":         {"мавик", "квадрокоптер", "дрон"},
		"мавики":        {"mavic", "мавик", "дроны", "квадрокоптеры"},
		"дрон":          {"квадрокоптер", "мавик", "mavic"},
		"квадрокоптер":  {"дрон", "мавик", "mavic"},
		"вэнди":         {"венди", "wendy", "team wendy", "тим венди"},
		"венди":         {"вэнди", "wendy", "team wendy", "тим венди"},
		"wendy":         {"венди", "вэнди", "тим венди", "team wendy"},
		"тим":           {"team"},
		"team":          {"тим"},
		"шлем":          {"каска", "helmet", "шлемы", "каски"},
		"шлемы":         {"каски", "шлем", "helmet"},
		"каска":         {"шлем", "helmet", "каски", "шлемы"},
		"каски":         {"шлемы", "каска", "шлем"},
		"плита":         {"бронеплита", "бронепластина", "керамика", "plate", "плиты"},
		"плиты":         {"бронеплиты", "бронепластины", "керамика", "плита"},
		"бронеплита":    {"плита", "бронепластина", "керамика", "бронеплиты"},
		"бронеплиты":    {"плиты", "бронепластины", "керамика", "бронеплита"},
		"керамика":      {"плиты", "бронеплиты", "плита", "бронеплита"},
		"броник":        {"бронежилет", "плитник"},
		"бронежилет":    {"броник", "плитник", "разгрузка"},
		"плитник":       {"бронежилет", "броник", "plate carrier"},
		"plate carrier": {"плитник", "бронежилет"},
		"пнв":           {"pnv", "ночное видение", "ночник"},
		"ночник":        {"пнв", "ночное видение"},
		"радиостанция":  {"радио", "radio", "рация", "рации", "радиостанции"},
		"радиостанции":  {"рации", "радиостанция", "рация"},
		"рация":         {"радиостанция", "radio", "рации", "радиостанции"},
		"рации":         {"радиостанции", "рация", "радиостанция"},
		"tyt":           {"тут", "рация", "радиостанция"},
		"тут":           {"tyt"},
		"dji":           {"дджи", "мавик", "mavic", "дрон", "квадрокоптер"},
		"акб":           {"аккумулятор", "батарея"},
		"батарея":       {"аккумулятор", "акб"},
		"аккумулятор":   {"акб", "батарея"},
		"бр5":           {"br5", "nij iv", "плиты", "керамика"},
	}

	if syns, ok := synonyms[query]; ok {
		for _, s := range syns {
			variationsMap[s] = true
		}
	}

	words := strings.Fields(query)
	for i, w := range words {
		if syns, ok := synonyms[w]; ok {
			for _, s := range syns {
				newWords := make([]string, len(words))
				copy(newWords, words)
				newWords[i] = s
				variationsMap[strings.Join(newWords, " ")] = true
			}
		}
	}

	var result []string
	for k := range variationsMap {
		if k != "" {
			result = append(result, k)
		}
	}
	return result
}

// SearchProducts выполняет полнотекстовый поиск по товарам
func SearchProducts(queryStr string) ([]models.Product, error) {
	var products []models.Product

	if strings.TrimSpace(queryStr) == "" {
		return products, nil
	}

	variations := getSearchVariations(queryStr)

	// Строим динамический запрос
	query := `
		SELECT DISTINCT
			p.id, p.name, p.category_id, c.name as category_name,
			p.price, p.old_price, p.badge, p.is_new, 
			p.images, p.documents, p.description, p.specs, p.created_at, p.views
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE `

	var args []interface{}
	var conditions []string

	for i, v := range variations {
		// Для каждого варианта генерируем по 3 плейсхолдера: $1, $2, $3 и т.д.
		idx := i * 3
		condition := fmt.Sprintf(`(
			(to_tsvector('russian', p.name || ' ' || coalesce(p.description, '')) @@ plainto_tsquery('russian', $%d))
			OR p.name ILIKE $%d 
			OR p.description ILIKE $%d
		)`, idx+1, idx+2, idx+3)
		conditions = append(conditions, condition)

		args = append(args, v, "%"+v+"%", "%"+v+"%")
	}

	query += strings.Join(conditions, " OR ")
	query += " ORDER BY p.id"

	err := database.DB.Select(&products, query, args...)
	if err != nil {
		return nil, err
	}

	for i := range products {
		products[i].ProcessNullFields()
	}

	return products, nil
}

// IncrementProductViews атомарно увеличивает счетчик просмотров
func IncrementProductViews(id int) error {
	query := `UPDATE products SET views = views + 1 WHERE id = $1`
	_, err := database.DB.Exec(query, id)
	return err
}

// GetTopViewedProducts возвращает самые популярные товары
func GetTopViewedProducts(limit int) ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT 
			p.id, p.name, p.category_id, c.name as category_name,
			p.price, p.views, p.images
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		ORDER BY p.views DESC
		LIMIT $1
	`
	err := database.DB.Select(&products, query, limit)
	return products, err
}

var enToRu = map[rune]rune{
	'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
	'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
	'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю', '/': '.',
	'Q': 'Й', 'W': 'Ц', 'E': 'У', 'R': 'К', 'T': 'Е', 'Y': 'Н', 'U': 'Г', 'I': 'Ш', 'O': 'Щ', 'P': 'З', '{': 'Х', '}': 'Ъ',
	'A': 'Ф', 'S': 'Ы', 'D': 'В', 'F': 'А', 'G': 'П', 'H': 'Р', 'J': 'О', 'K': 'Л', 'L': 'Д', ':': 'Ж', '"': 'Э',
	'Z': 'Я', 'X': 'Ч', 'C': 'С', 'V': 'М', 'B': 'И', 'N': 'Т', 'M': 'Ь', '<': 'Б', '>': 'Ю', '?': ',',
}

var ruToEn = make(map[rune]rune)

func init() {
	for k, v := range enToRu {
		ruToEn[v] = k
	}
}

// switchLayout меняет раскладку текста (RU <-> EN)
func switchLayout(input string) string {
	res := make([]rune, 0, len(input))
	for _, r := range input {
		if val, ok := enToRu[r]; ok {
			res = append(res, val)
		} else if val, ok := ruToEn[r]; ok {
			res = append(res, val)
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}
