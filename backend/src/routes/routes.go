package routes

import (
	"go-admin/src/controllers"
	"go-admin/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func Setup(router *fiber.App) {

	router.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!🥳")
	})

	api := router.Group("/api")

	//* Routes groups
	adminRoutes := api.Group("/admin")
	ambassadorRoutes := api.Group("/ambassador")
	productRoutes := api.Group("/products")
	userRoutes := api.Group("/users")
	checkoutRoutes := api.Group("/checkout")

	//* Public Routes
	adminRoutes.Post("/register", controllers.Register)
	adminRoutes.Post("/login", controllers.Login)

	ambassadorRoutes.Post("/register", controllers.Register)
	ambassadorRoutes.Post("/login", controllers.Login)

	//* Private Routes
	api.Use(middlewares.IsAuthenticated)

	api.Post("/logout", controllers.Logout)

	adminRoutes.Get("/me", controllers.User)

	ambassadorRoutes.Get("/", controllers.GetAmbassadors)
	ambassadorRoutes.Get("/me", controllers.User)
	ambassadorRoutes.Post("/link", controllers.CreateLink)
	ambassadorRoutes.Get("/stats", controllers.Stats)
	ambassadorRoutes.Get("/rankings", controllers.Rankings)

	userRoutes.Patch("/info", controllers.UpdateInfo)
	userRoutes.Patch("/password", controllers.UpdatePassword)
	userRoutes.Get("/:id/links", controllers.GetLinks)

	productRoutes.Get("/", controllers.GetProducts)
	productRoutes.Post("/", controllers.CreateProduct)
	productRoutes.Get("/frontend", controllers.ProductsFrontend)
	productRoutes.Get("/backend", controllers.ProductsBackend)
	productRoutes.Get("/:id", controllers.GetProduct)
	productRoutes.Put("/:id", controllers.UpdateProduct)
	productRoutes.Delete("/:id", controllers.DeleteProduct)

	api.Get("/orders", controllers.GetOrders)

	checkoutRoutes.Get("links/:code", controllers.GetLink)
	checkoutRoutes.Post("/orders", controllers.CreateOrder)
	checkoutRoutes.Post("/orders/confirm", controllers.CompleteOrder)
}
