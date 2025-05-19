package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// User представляет пользователя системы
type User struct {
	Username string
	Password string
}

// Middleware для базовой аутентификации
func BasicAuth(username, password string, enabled bool) gin.HandlerFunc {
	// Если аутентификация выключена, просто пропускать запросы
	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Получение заголовка авторизации
		auth := c.GetHeader("Authorization")

		// Проверка наличия заголовка
		if auth == "" {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Проверка формата
		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Декодирование и проверка учетных данных
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Разбор учетных данных (формат: username:password)
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			c.Header("WWW-Authenticate", "Basic realm=TeleOko")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Если все проверки пройдены
		c.Next()
	}
}
