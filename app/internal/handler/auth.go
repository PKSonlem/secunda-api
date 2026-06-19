package handler

import (
	"net/http"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	svc domain.AuthService
}

func NewAuthHandler(svc domain.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.svc.Register(c.Request().Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.svc.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
