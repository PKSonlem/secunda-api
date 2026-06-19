package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const ctxUserID = "userID"

type tokenValidator interface {
	ValidateToken(token string) (int64, error)
}

func JWT(svc tokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}
			userID, err := svc.ValidateToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			c.Set(ctxUserID, userID)
			return next(c)
		}
	}
}

func UserIDFromCtx(c echo.Context) int64 {
	id, _ := c.Get(ctxUserID).(int64)
	return id
}
