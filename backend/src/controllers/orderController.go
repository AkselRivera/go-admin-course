package controllers

import (
	"context"
	"fmt"
	"go-admin/src/database"
	"go-admin/src/models"
	"net/smtp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

func GetOrders(c *fiber.Ctx) error {
	var orders []models.Order

	database.DB.Preload("OrderItems").Find(&orders)

	for i, order := range orders {
		orders[i].Name = order.FullName()
		orders[i].Total = order.GetTotal()
	}

	return c.JSON(orders)
}

type CreateOrderRequest struct {
	Code      string           `json:"code"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Address   string           `json:"address"`
	Country   string           `json:"country"`
	City      string           `json:"city"`
	Zip       string           `json:"zip"`
	Products  []map[string]int `json:"products"`
}

func CreateOrder(c *fiber.Ctx) error {

	var request CreateOrderRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	link := models.Link{
		Code: request.Code,
	}

	database.DB.Preload("User").First(&link)

	if link.Id == 0 {
		return c.Status(400).JSON("Invalid link")
	}

	order := models.Order{
		Code:            link.Code,
		UserId:          link.UserId,
		AmbassadorEmail: link.User.Email,
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		Email:           request.Email,
		Address:         request.Address,
		Country:         request.Country,
		City:            request.City,
		Zip:             request.Zip,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{}

	for _, requestProduct := range request.Products {
		id := requestProduct["product_id"]
		product := models.Product{}
		product.Id = id

		database.DB.First(&product)

		total := product.Price * float64(requestProduct["quantity"])

		item := models.OrderItem{
			OrderId:           order.Id,
			ProductTitle:      product.Title,
			Price:             product.Price,
			Quantity:          int(requestProduct["quantity"]),
			AmbassadorRevenue: 0.1 * total,
			AdminRevenue:      0.9 * total,
		}

		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		price := fmt.Sprintf("%.2f", product.Price)

		fixedPrice, _ := strconv.ParseFloat(price, 64)

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String("mxn"),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name:        stripe.String(product.Title),
					Description: stripe.String(product.Description),
					Images: []*string{
						stripe.String(product.Image),
					},
				},
				UnitAmountDecimal: stripe.Float64(100 * fixedPrice),
			},
			Quantity: stripe.Int64(int64(requestProduct["quantity"])),
		})

	}

	stripe.Key = "sk_test_51P3suyDLhhQ55eUjBcRjEQcvzQxuGQ2y2mnLWfO27092mhIHbXr3e53FWmEQ7IsPrwvRcs24YSrDHNWycNReKsPX00OimcHCc0"

	params := stripe.CheckoutSessionParams{
		SuccessURL: stripe.String("http://localhost:3000/success?source={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String("http://localhost:3000/error"),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: lineItems,
		Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
	}

	source, err := session.New(&params)

	if err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	order.TransactionId = source.ID

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	tx.Commit()

	return c.Status(201).JSON(source)
}

func CompleteOrder(c *fiber.Ctx) error {
	var request struct {
		Source string `json:"source"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(err.Error())
	}

	order := models.Order{}

	database.DB.Preload("OrderItems").First(&order, &models.Order{
		TransactionId: request.Source,
	})

	if order.Id == 0 {
		return c.Status(404).JSON(fiber.Map{
			"message": "Order not found",
		})
	}

	order.Complete = true

	database.DB.Save(&order)

	go func(order models.Order) {

		ambassadorRevenue := 0.0
		adminRevenue := 0.0

		for _, orderItem := range order.OrderItems {
			ambassadorRevenue += orderItem.AmbassadorRevenue
			adminRevenue += orderItem.AdminRevenue
		}

		user := models.User{}
		user.Id = order.UserId

		database.DB.First(&user)

		database.Cache.ZIncrBy(context.Background(), "rankings", ambassadorRevenue, user.FullName())

		ambassadorMessage := fmt.Sprintf("You earned $%.2f revenue from the link #%s \r\n", ambassadorRevenue, order.Code)

		msg := []byte(fmt.Sprintf("From: Gopher <no-reply@localhost.com>\r\n"+
			"To: %s\r\n"+
			"Subject: Order completed Gopher!\r\n"+
			"\r\n"+
			"%s.\r\n", order.AmbassadorEmail, ambassadorMessage))

		smtp.SendMail("host.docker.internal:1025", nil, "no-reply@localhost.com", []string{order.AmbassadorEmail}, msg)

		adminMessage := fmt.Sprintf("Order #%s has been completed. You earned $%.2f \r\n", order.Code, adminRevenue)

		msg = []byte(fmt.Sprintf("From: Gopher <no-reply@localhost.com>\r\n"+
			"To: %s\r\n"+
			"Subject: Chief Gopher, Order completed!\r\n"+
			"\r\n"+
			"%s.\r\n", order.AmbassadorEmail, adminMessage))

		smtp.SendMail("host.docker.internal:1025", nil, "no-reply@localhost.com", []string{order.AmbassadorEmail, "moralesaksel@gmail.com"}, msg)

	}(order)

	return c.JSON(fiber.Map{
		"message": "Order completed",
	})
}
