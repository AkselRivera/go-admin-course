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
		product := models.Product{
			Title:       faker.Name(),
			Description: faker.Paragraph(),
			Image:       faker.URL(),
			Price:       float64(rand.Float64() + 100),
		}

		database.DB.Create(&product)
	}
}
