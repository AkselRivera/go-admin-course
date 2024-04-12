package controllers

import (
	"go-admin/src/database"
	"go-admin/src/middlewares"
	"go-admin/src/models"
	"strconv"

	"github.com/bxcodec/faker/v4"
	"github.com/gofiber/fiber/v2"
)

func GetLinks(c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(400).JSON("Please enter a valid id")
	}

	var links []models.Link
	database.DB.Where("user_id = ?", id).Find(&links)

	for i, link := range links {
		var order []models.Order

		database.DB.Where("code = ? and complete = true", link.Code).First(&order)

		if len(order) > 0 {
			links[i].Orders = order
		}
	}

	return c.JSON(links)
}

func GetLink(c *fiber.Ctx) error {

	code := c.Params("code")

	links := models.Link{
		Code: code,
	}

	database.DB.Preload("User").Preload("Products").First(&links)

	return c.JSON(links)
}

type CreateLinkRequest struct {
	Product []int `json:"products"`
}

func CreateLink(c *fiber.Ctx) error {
	var request CreateLinkRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	id, _ := middlewares.GetUserId(c)

	link := models.Link{
		Code:   faker.Word(),
		UserId: id,
	}

	for _, id := range request.Product {
		product := models.Product{}
		product.Id = id
		link.Products = append(link.Products, product)
	}

	database.DB.Create(&link).Preload("Products")

	return c.Status(201).JSON(link)
}

func Stats(c *fiber.Ctx) error {
	id, _ := middlewares.GetUserId(c)

	var links []models.Link
	database.DB.Find(&links, models.Link{
		UserId: id,
	})

	var result []interface{}

	for _, link := range links {
		var orders []models.Order
		database.DB.Preload("OrderItems").Find(&orders, models.Order{
			Code:     link.Code,
			Complete: true,
		})

		revenue := 0.0

		for _, order := range orders {
			revenue += order.GetTotal()
		}

		result = append(result, fiber.Map{
			"code":    link.Code,
			"count":   len(orders),
			"revenue": revenue,
		})
	}

	return c.JSON(result)
}
