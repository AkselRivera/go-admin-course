package controllers

import (
	"context"
	"encoding/json"
	"go-admin/src/database"
	"go-admin/src/models"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetProducts(c *fiber.Ctx) error {

	var products []models.Product
	database.DB.Find(&products)

	return c.JSON(products)
}

func CreateProduct(c *fiber.Ctx) error {

	var product models.Product

	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(err.Error())
	}
	database.DB.Create(&product)

	go database.ClearCache("products_frontend", "products_backend")

	return c.JSON(product)
}

func GetProduct(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON("Please enter a valid id")
	}

	var product models.Product

	product.Id = id

	database.DB.Find(&product)

	return c.JSON(product)
}

func UpdateProduct(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON("Please enter a valid id")
	}

	product := models.Product{}
	product.Id = id

	if err := c.BodyParser(&product); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	database.DB.Model(&product).Updates(&product)

	go database.ClearCache("products_frontend", "products_backend")

	return c.JSON(product)
}

func DeleteProduct(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON("Please enter a valid id")
	}

	product := models.Product{}
	product.Id = id

	database.DB.Delete(&product)

	go database.ClearCache("products_frontend", "products_backend")

	return nil
}

func ProductsFrontend(c *fiber.Ctx) error {

	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_frontend").Result()

	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)

		if err != nil {
			c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something went wrong when trying to cache products", "data": err.Error()})
		}

		database.Cache.Set(ctx, "products_frontend", bytes, 30*time.Minute)

	} else {
		json.Unmarshal([]byte(result), &products)
	}

	return c.JSON(products)
}

func ProductsBackend(c *fiber.Ctx) error {

	var products []models.Product
	var ctx = context.Background()

	result, err := database.Cache.Get(ctx, "products_backend").Result()

	if err != nil {
		database.DB.Find(&products)

		bytes, err := json.Marshal(products)

		if err != nil {
			c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something went wrong when trying to cache products", "data": err.Error()})
		}

		database.Cache.Set(ctx, "products_backend", bytes, 30*time.Minute)

	} else {
		json.Unmarshal([]byte(result), &products)
	}

	var searchedProducts []models.Product

	if s := c.Query("s"); s != "" {
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Title), strings.ToLower(s)) || strings.Contains(strings.ToLower(product.Description), strings.ToLower(s)) {
				searchedProducts = append(searchedProducts, product)
			}
		}
		products = searchedProducts
	}

	if sortParam := c.Query("sort"); sortParam != "" {
		sortLower := strings.ToLower(sortParam)

		switch sortLower {
		case "asc":
			sort.Slice(products, func(i, j int) bool {
				return products[i].Price < products[j].Price
			})

		case "desc":
			sort.Slice(products, func(i, j int) bool {
				return products[i].Price > products[j].Price
			})

		default:
			return c.Status(400).JSON("Invalid sort param")

		}
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage := 20
	total := len(products)
	// limit := c.Query("limit", strconv.Itoa(10))
	// offset := (pageInt - 1) * limitInt

	var data []models.Product = products

	if total <= page*perPage && total >= (page-1)*perPage {
		data = data[(page-1)*perPage : total]
	} else if total >= page*perPage {
		data = data[(page-1)*perPage : page*perPage]
	} else {
		data = []models.Product{}
	}

	return c.JSON(fiber.Map{
		"data":      data,
		"total:":    total,
		"page":      page,
		"last_page": total/perPage + 1,
	})
}
