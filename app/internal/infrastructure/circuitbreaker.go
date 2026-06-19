package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PKSonlem/testtask-secunda-api/internal/config"
	"github.com/sony/gobreaker"
)

type EmailClient struct {
	url     string
	breaker *gobreaker.CircuitBreaker
	http    *http.Client
}

// NewEmailCircuitBreaker размыкает цепь после 3 последовательных ошибок и держит её разомкнутой
// 30 секунд. MaxRequests=3 — число пробных запросов в полуоткрытом состоянии.
func NewEmailCircuitBreaker(cfg config.EmailConfig) *EmailClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "email-service",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			return c.ConsecutiveFailures > 3
		},
	})

	return &EmailClient{
		url:     cfg.URL,
		breaker: cb,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// SendInvite считает ошибкой только ответы 5xx — 4xx не трипают circuit breaker,
// так как это проблема запроса, а не доступности сервиса.
func (e *EmailClient) SendInvite(ctx context.Context, email, teamName string) error {
	_, err := e.breaker.Execute(func() (interface{}, error) {
		body, _ := json.Marshal(map[string]string{"email": email, "team": teamName})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url+"/invite", strings.NewReader(string(body)))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := e.http.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("email service error: %d", resp.StatusCode)
		}

		return nil, nil
	})

	return err
}
