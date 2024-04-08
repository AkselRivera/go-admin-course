package main

import (
	"go-admin/src/database"
	"go-admin/src/models"
	"math/rand"

	"github.com/bxcodec/faker/v4"
)

func main() {
	database.Connect()

	for i := 0; i < 30; i++ {
		var orderItem []models.OrderItem

		for j := 0; j < rand.Intn(5)+1; j++ {
			price := rand.Float64() + 90
			qty := rand.Intn(5) + 1

			orderItem = append(orderItem, models.OrderItem{
				ProductTitle:      faker.Name(),
				Price:             price,
				Quantity:          qty,
				AdminRevenue:      float64(qty) * price * 0.9,
				AmbassadorRevenue: float64(qty) * price * 0.1,
			})

			database.DB.Create(&models.Order{
				UserId:          rand.Intn(30) + 1,
				Code:            faker.Word(),
				AmbassadorEmail: faker.Email(),
				FirstName:       faker.FirstName(),
				LastName:        faker.LastName(),
				Email:           faker.Email(),
				Complete:        true,
				OrderItems:      orderItem,
			})
		}
	}

}
