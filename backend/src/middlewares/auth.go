package middlewares

import (
	"go-admin/src/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const SecretKey = "secret"

type CustomClaims struct {
	FistName string `json:"first_name"`
	Scope    string
	jwt.RegisteredClaims
}

func IsAuthenticated(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthenticated",
		})
	}

	claims := token.Claims.(*CustomClaims)
	isAmbassador := strings.Contains(c.Path(), "/api/ambassador")

	if (claims.Scope == "admin" && isAmbassador) || (claims.Scope == "ambassador" && !isAmbassador) {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	return c.Next()
}

func GetUserId(c *fiber.Ctx) (int, error) {

	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	if err != nil {
		return 0, err
	}

	payload := token.Claims.(*CustomClaims)

	id, _ := strconv.Atoi(payload.RegisteredClaims.Subject)

	return id, nil

}

func GenerateJWT(user models.User, scope string) (string, error) {

	payload := CustomClaims{
		FistName: user.FirstName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			Subject:   strconv.Itoa(int(user.Id)),
		},
		Scope: scope,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(SecretKey))

}
