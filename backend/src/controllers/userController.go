package controllers

import (
	"context"
	"go-admin/src/database"
	"go-admin/src/models"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

func GetAmbassadors(c *fiber.Ctx) error {
	var users []models.User
	database.DB.Where("is_ambassador = true").Find(&users)
	return c.JSON(users)
}

func Rankings(c *fiber.Ctx) error {

	// var users []models.User

	// database.DB.Order("revenue desc").Find(&users, models.User{IsAmbassador: true})
	// database.DB.Find(&users, models.User{IsAmbassador: true})

	// var result []interface{}

	// for _, user := range users {

	// 	ambassador := models.Ambassador(user)
	// 	ambassador.CalculateRevenue(database.DB)

	// 	result = append(result, fiber.Map{
	// 		user.FullName(): ambassador.Revenue,
	// 	})
	// }

	rankings, err := database.Cache.ZRevRangeByScoreWithScores(context.Background(), "rankings", &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}).Result()

	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.Status(500).JSON(fiber.Map{
			"message": "could not fetch data",
		})
	}

	result := make(map[string]float64)

	for _, ranking := range rankings {
		result[ranking.Member.(string)] = ranking.Score
	}

	return c.JSON(result)

}
