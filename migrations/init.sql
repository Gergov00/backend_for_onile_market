-- =====================================================
-- MILITARY SHOP - ПОЛНАЯ МИГРАЦИЯ БАЗЫ ДАННЫХ
-- Версия: 1.0
-- =====================================================
-- Этот файл содержит ВСЮ структуру БД и начальные данные
-- Для чистой установки: выполните этот файл целиком
-- =====================================================

-- Удаляем таблицы если существуют (для чистой установки)
DROP TABLE IF EXISTS verification_codes CASCADE;
DROP TABLE IF EXISTS cart_items CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS site_stats CASCADE;

-- =====================================================
-- КАТЕГОРИИ
-- =====================================================
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    image VARCHAR(500),
    count INTEGER DEFAULT 0
);

-- =====================================================
-- ТОВАРЫ (с images массивом)
-- =====================================================
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    price DECIMAL(10,2) NOT NULL,
    old_price DECIMAL(10,2),
    badge VARCHAR(50),
    is_new BOOLEAN DEFAULT FALSE,
    images TEXT[],
    documents TEXT[],
    description TEXT,
    specs JSONB,
    views INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ПОЛЬЗОВАТЕЛИ (с phone_verified)
-- =====================================================
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100),
    phone VARCHAR(20),
    phone_verified BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Уникальный индекс для телефона
CREATE UNIQUE INDEX idx_users_phone_unique 
ON users(phone) WHERE phone IS NOT NULL AND phone != '';

-- =====================================================
-- ЗАКАЗЫ (ОБНОВЛЕНО: UserID может быть NULL, добавлен taken_by)
-- =====================================================
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50) NOT NULL,
    messenger VARCHAR(50) NOT NULL,
    comment TEXT,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'new',
    address TEXT,
    taken_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ПОЗИЦИИ ЗАКАЗА
-- =====================================================
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

-- =====================================================
-- КОРЗИНА
-- =====================================================
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 1,
    UNIQUE(user_id, product_id)
);

-- =====================================================
-- КОДЫ ВЕРИФИКАЦИИ ТЕЛЕФОНА
-- =====================================================
CREATE TABLE verification_codes (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20),
    email VARCHAR(255),
    code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INTEGER DEFAULT 0,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_verify_contact CHECK (phone IS NOT NULL OR email IS NOT NULL)
);

-- =====================================================
-- СТАТИСТИКА ПОСЕЩЕНИЙ
-- =====================================================
CREATE TABLE site_stats (
    id SERIAL PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    visits INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ГАЛЕРЕЯ (фото и видео производства/завода)
-- =====================================================
CREATE TABLE gallery_items (
    id SERIAL PRIMARY KEY,
    type VARCHAR(10) NOT NULL CHECK (type IN ('photo', 'video')),
    url VARCHAR(500) NOT NULL,
    title VARCHAR(255),
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- ИНДЕКСЫ
-- =====================================================
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_cart_user ON cart_items(user_id);
CREATE INDEX idx_verification_phone ON verification_codes(phone);
CREATE INDEX idx_verification_email ON verification_codes(email);
CREATE INDEX idx_verification_expires ON verification_codes(expires_at);
CREATE INDEX idx_gallery_type ON gallery_items(type);
CREATE INDEX idx_gallery_sort ON gallery_items(sort_order);

-- =====================================================
-- НАЧАЛЬНЫЕ ДАННЫЕ: КАТЕГОРИИ
-- =====================================================
INSERT INTO categories (name, slug, description, image, count) VALUES
('Одежда', 'odezhda', 'Форменная одежда и камуфляж для военных и профессионалов', '/images/categories/clothes.jpg', 156),
('Обувь', 'obuv', 'Берцы, ботинки и специальная обувь для любых условий', '/images/categories/boots.jpg', 89),
('Рюкзаки', 'ryukzaki', 'Тактические рюкзаки и сумки для переноски снаряжения', '/images/categories/backpacks.jpg', 67),
('Снаряжение', 'snaryazhenie', 'Разгрузки, ремни, подсумки и другое тактическое снаряжение', '/images/categories/gear.jpg', 234),
('Оптика', 'optika', 'Бинокли, прицелы, ПНВ и другие оптические приборы', '/images/categories/optics.jpg', 78),
('Туризм', 'turizm', 'Снаряжение для походов, выживания и активного отдыха', '/images/categories/camping.jpg', 145);

-- =====================================================
-- НАЧАЛЬНЫЕ ДАННЫЕ: ТОВАРЫ
-- =====================================================
INSERT INTO products (name, category_id, price, old_price, badge, is_new, images, description, specs) VALUES
(
    'Армейский рюкзак десантника 60л',
    3, 5500, 6900, 'ХИТ', FALSE,
    ARRAY['/images/products/backpack_1.png', '/images/products/backpack_2.png'],
    'Прочный тактический рюкзак объёмом 60 литров. Выполнен из водоотталкивающей ткани повышенной прочности.',
    '["Объём: 60 литров", "Материал: Cordura 1000D", "Водоотталкивающее покрытие", "Вес: 1.8 кг"]'
),
(
    'Берцы военные уставные',
    2, 7990, NULL, 'NEW', TRUE,
    ARRAY['/images/products/boots_1.png', '/images/products/boots_2.png'],
    'Уставные армейские берцы из натуральной кожи. Подошва устойчива к агрессивным средам.',
    '["Материал: натуральная кожа", "Подошва: полиуретан/резина", "Размеры: 39-47"]'
),
(
    'Бронежилет тактический 4 класс',
    4, 45000, 52000, '-13%', FALSE,
    ARRAY['/images/products/vest_1.png'],
    'Современный бронежилет 4 класса защиты. Обеспечивает защиту от пистолетных пуль и осколков.',
    '["Класс защиты: 4", "Площадь защиты: 48 дм²", "Вес без пластин: 2.5 кг"]'
),
(
    'Форма полевая ВКПО',
    1, 8900, NULL, NULL, FALSE,
    ARRAY['/images/products/uniform_1.png'],
    'Полевая форма нового образца из комплекта ВКПО.',
    '["Материал: рип-стоп 50/50", "Влагоотталкивающая пропитка", "Размеры: 44-58"]'
),
(
    'Штык-нож армейский 6х5',
    4, 4800, 5900, 'ХИТ', FALSE,
    ARRAY['/images/products/knife_1.png'],
    'Классический армейский штык-нож 6х5. Универсальный инструмент для полевых условий.',
    '["Длина клинка: 160 мм", "Сталь: 65Х13", "Рукоять: полиамид"]'
),
(
    'Прибор ночного видения ПНВ-57',
    5, 89000, NULL, 'NEW', TRUE,
    ARRAY['/images/products/nvg_1.png'],
    'Профессиональный прибор ночного видения поколения 2+.',
    '["Поколение: 2+", "Увеличение: 1x", "Дальность: до 200м"]'
),
(
    'Каска армейская 6Б47',
    4, 15900, 19000, '-16%', FALSE,
    ARRAY['/images/products/helmet_1.png'],
    'Современный армейский шлем 6Б47. Обеспечивает защиту от осколков и рикошетов.',
    '["Класс защиты: Бр2", "Материал: арамид", "Вес: 1.1 кг"]'
),
(
    'Разгрузочный жилет 6Ш117',
    4, 12990, NULL, NULL, FALSE,
    ARRAY['/images/products/vest_2.png'],
    'Тактический разгрузочный жилет для переноски боекомплекта и снаряжения.',
    '["Система: MOLLE", "Материал: Cordura 500D", "Вес: 1.3 кг"]'
);
