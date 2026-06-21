package handlertest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/PKSonlem/testtask-secunda-api/internal/handler"
	"github.com/PKSonlem/testtask-secunda-api/tests/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	return e
}

// call invokes a handler and applies echo's error handler so the recorder always has a status.
func call(e *echo.Echo, h echo.HandlerFunc, c echo.Context) {
	if err := h(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
}

func TestAuthHandler_Register(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		svc := &mocks.AuthSvc{
			RegisterFn: func(_ context.Context, email, _, name string) (*domain.User, error) {
				return &domain.User{ID: 1, Email: email, Name: name}, nil
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`{"email":"test@test.com","password":"pass","name":"Test"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewAuthHandler(svc).Register, e.NewContext(req, rec))

		assert.Equal(t, http.StatusCreated, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "test@test.com", resp["email"])
	})

	t.Run("service_error", func(t *testing.T) {
		svc := &mocks.AuthSvc{
			RegisterFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
				return nil, assert.AnError
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/register",
			strings.NewReader(`{"email":"x@x.com","password":"pass","name":"X"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewAuthHandler(svc).Register, e.NewContext(req, rec))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	e := newEcho()

	t.Run("success", func(t *testing.T) {
		svc := &mocks.AuthSvc{
			LoginFn: func(_ context.Context, _, _ string) (string, error) {
				return "jwt.token.here", nil
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`{"email":"test@test.com","password":"pass"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewAuthHandler(svc).Login, e.NewContext(req, rec))

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "jwt.token.here", resp["token"])
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		svc := &mocks.AuthSvc{
			LoginFn: func(_ context.Context, _, _ string) (string, error) {
				return "", domain.ErrUnauthorized
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/login",
			strings.NewReader(`{"email":"x@x.com","password":"wrong"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		call(e, handler.NewAuthHandler(svc).Login, e.NewContext(req, rec))

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
