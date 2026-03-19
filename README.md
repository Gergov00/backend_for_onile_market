# Military Shop API (Backend)

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Framework-Gin-008ECF?style=flat)](https://github.com/gin-gonic/gin)
[![PostgreSQL](https://img.shields.io/badge/Database-PostgreSQL-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)

Современный и высокопроизводительный RESTful API для онлайн-магазина тактического снаряжения. Написан на Go с использованием фреймворка Gin и архитектуры Clean Architecture-ish.

---

## Основные возможности

*   **Безопасность и Аутентификация**:
    *   Регистрация по Email с подтверждением.
    *   JWT-авторизация для пользователей.
    *   Role-based Access Control (RBAC): разделение прав пользователей и администраторов.
*   **Электронная коммерция**:
    *   Управление корзиной покупок.
    *   Система оформления и отслеживания заказов.
    *   Интеграция с ботом для обработки заказов администраторами.
*   **Каталог товаров**:
    *   Динамические категории и подкатегории.
    *   Поиск товаров с поддержкой синонимов.
    *   Галерея изображений для товаров и страниц.
*   **Аналитика и Администрирование**:
    *   Панель статистики для администраторов.
    *   Управление контентом статических страниц (Landing Pages).
    *   Генерация динамической Sitemap для SEO.
*   **Интеграции**:
    *   **Notificore**: SMS-уведомления (с mock-режимом для разработки).
    *   **SMTP (Beget)**: Почтовые рассылки и восстановление паролей.

---

## Стек технологий

*   **Язык**: [Go (Golang) 1.24](https://golang.org/)
*   **HTTP Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
*   **База данных**: [PostgreSQL](https://www.postgresql.org/)
*   **Библиотеки для работы с БД**: [sqlx](https://github.com/jmoiron/sqlx), `lib/pq`
*   **CI/CD & Deployment**: Docker, Docker Compose
*   **Валидация**: `go-playground/validator`
*   **Логирование**: Встроенный Gin Logger + Go log

---

## Структура проекта

```text
.
├── cmd/
│   └── server/          # Точка входа в приложение (main.go)
├── internal/
│   ├── database/        # Инициализация и подключение к БД
│   ├── handlers/        # Обработчики HTTP-запросов (контроллеры)
│   ├── middleware/      # Промежуточное ПО (Auth, Admin, CORS, Recovery)
│   ├── models/          # Описание структур данных и сущностей БД
│   ├── repository/      # Слой работы с данными (SQL запросы)
│   └── services/        # Бизнес-логика приложения
├── migrations/          # SQL файлы для миграции структуры БД
├── uploads/             # Статические файлы (изображения товаров)
├── Dockerfile           # Конфигурация образа Docker
├── .env.example         # Пример конфигурации переменных окружения
└── go.mod               # Зависимости проекта
```

---

## Настройка и запуск

### 1. Подготовка окружения
Создайте файл `.env` в корневой директории и заполните его по примеру:

```bash
cp .env.example .env
```

Основные параметры:
- `DB_*`: Настройки подключения к PostgreSQL.
- `JWT_SECRET`: Секретный ключ для подписи токенов (измените в продакшене!).
- `NOTIFICORE_API_KEY`: API ключ для SMS (если пусто — используется консольный вывод).
- `SMTP_*`: Настройки почтового сервера.

### 2. Запуск через Docker (рекомендуется)
Самый простой способ запустить проект вместе с базой данных:

```bash
docker-compose up --build
```

### 3. Локальный запуск
Если вы хотите запустить проект без Docker:

1.  Убедитесь, что у вас установлен **Go 1.24+** и запущен **PostgreSQL**.
2.  Загрузите зависимости:
    ```bash
    go mod download
    ```
3.  Создайте базу данных и примените миграции (из папки `/migrations`).
4.  Запустите сервер:
    ```bash
    go run cmd/server/main.go
    ```

---

## API Эндпоинты (Кратко)

| Метод | Путь | Описание | Доступ |
| :--- | :--- | :--- | :--- |
| **POST** | `/api/auth/register/email` | Регистрация | Public |
| **POST** | `/api/auth/login` | Вход и получение JWT | Public |
| **GET** | `/api/products` | Список товаров (поиск, фильтры) | Public |
| **GET** | `/api/cart` | Просмотр корзины | User |
| **POST** | `/api/orders` | Оформление заказа | User |
| **GET** | `/api/admin/stats` | Статистика магазина | Admin |
| **GET** | `/api/bot/orders` | Заказы для внешнего бота | Bot/Admin |

---

## Тестирование
Для запуска тестов используйте стандартную команду Go:

```bash
go test ./...
```

---

## Контакты
Если у вас есть вопросы по проекту или предложения по улучшению — создавайте Issue или Pull Request.

---
*Разработано для проекта Military Shop.*
