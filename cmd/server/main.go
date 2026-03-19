package main

import (
	"log"
	"os"

	"military-shop/internal/database"
	"military-shop/internal/handlers"
	"military-shop/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer database.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/api/orders", "/api/bot/orders"},
	}))
	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "https://russian-armo.org", "https://www.russian-armo.org"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Static("/uploads", "./uploads")
	r.MaxMultipartMemory = 500 << 20

	api := r.Group("/api")
	{
		api.POST("/upload", middleware.AuthRequired(), handlers.UploadImage)

		api.POST("/visit", handlers.RecordVisit)

		auth := api.Group("/auth")
		{
			auth.POST("/register/email", handlers.RegisterEmail) // [НОВОЕ]
			auth.POST("/verify-email", handlers.VerifyEmail)     // [НОВОЕ]
			auth.POST("/login", handlers.Login)
			auth.POST("/forgot-password", handlers.ForgotPassword)
			auth.POST("/reset-password", handlers.ResetPassword)
			auth.GET("/me", middleware.AuthRequired(), handlers.GetMe)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
		{
			admin.GET("/stats", handlers.GetAdminStats)
			admin.POST("/products", handlers.CreateProduct)
			admin.PUT("/products/:id", handlers.UpdateProduct)
			admin.DELETE("/products/:id", handlers.DeleteProduct)

			admin.POST("/categories", handlers.CreateCategory)
			admin.PUT("/categories/:id", handlers.UpdateCategory)
			admin.DELETE("/categories/:id", handlers.DeleteCategory)

			admin.POST("/gallery", handlers.CreateGalleryItem)
			admin.DELETE("/gallery/:id", handlers.DeleteGalleryItem)

			admin.PUT("/pages/:slug", handlers.UpdatePage)
		}

		api.GET("/categories", handlers.GetCategories)
		api.GET("/categories/:slug", handlers.GetCategoryBySlug)

		api.GET("/pages", handlers.GetPages)
		api.GET("/pages/:slug", handlers.GetPageBySlug)

		api.GET("/products", handlers.GetProducts)
		api.GET("/products/search", handlers.SearchProducts)
		api.GET("/products/:id", handlers.GetProductByID)

		api.GET("/sitemap", handlers.GenerateSitemap)

		api.GET("/gallery", handlers.GetGalleryItems)

		cart := api.Group("/cart")
		cart.Use(middleware.AuthRequired())
		{
			cart.GET("", handlers.GetCart)
			cart.POST("", handlers.AddToCart)
			cart.PUT("/:id", handlers.UpdateCartItem)
			cart.DELETE("/:id", handlers.RemoveFromCart)
			cart.DELETE("", handlers.ClearCart)
		}

		orders := api.Group("/orders")
		orders.Use(middleware.AuthRequired())
		{
			orders.GET("", handlers.GetUserOrders)
			orders.GET("/:id", handlers.GetOrderByID)
			orders.POST("", handlers.CreateOrder)
		}

		bot := api.Group("/bot")
		{
			bot.GET("/orders", handlers.GetBotOrders)
			bot.GET("/orders/:id", handlers.GetBotOrder)
			bot.POST("/orders/:id/take", handlers.BotTakeOrder)
			bot.POST("/orders/:id/status", handlers.BotUpdateStatus)
		}

	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
