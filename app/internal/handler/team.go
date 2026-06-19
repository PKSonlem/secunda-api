package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/middleware"
	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	svc domain.TeamService
}

func NewTeamHandler(svc domain.TeamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

func (h *TeamHandler) Create(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	team, err := h.svc.Create(c.Request().Context(), middleware.UserIDFromCtx(c), req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) List(c echo.Context) error {
	teams, err := h.svc.List(c.Request().Context(), middleware.UserIDFromCtx(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, teams)
}

func (h *TeamHandler) Invite(c echo.Context) error {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid team id")
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = h.svc.Invite(c.Request().Context(), middleware.UserIDFromCtx(c), teamID, req.Email)

	switch {
	case err == nil:
		return c.JSON(http.StatusOK, echo.Map{"ok": true})
	case errors.Is(err, domain.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
}
