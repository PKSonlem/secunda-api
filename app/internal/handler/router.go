package handler

import (
	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	mw "github.com/PKSonlem/testtask-secunda-api/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func NewRouter(
	authSvc domain.AuthService,
	teamSvc domain.TeamService,
	taskSvc domain.TaskService,
	rdb *redis.Client,
) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(mw.Logger())
	e.Use(mw.Metrics())

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	v1 := e.Group("/api/v1")

	auth := NewAuthHandler(authSvc)
	v1.POST("/register", auth.Register)
	v1.POST("/login", auth.Login)

	authed := v1.Group("")
	authed.Use(mw.JWT(authSvc))
	authed.Use(mw.RateLimit(rdb))

	teams := NewTeamHandler(teamSvc)
	authed.POST("/teams", teams.Create)
	authed.GET("/teams", teams.List)
	authed.POST("/teams/:id/invite", teams.Invite)

	tasks := NewTaskHandler(taskSvc)
	authed.POST("/tasks", tasks.Create)
	authed.GET("/tasks", tasks.List)
	authed.PUT("/tasks/:id", tasks.Update)
	authed.GET("/tasks/:id/history", tasks.GetHistory)

	return e
}
