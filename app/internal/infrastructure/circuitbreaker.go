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
