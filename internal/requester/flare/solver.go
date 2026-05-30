package flare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultURL     = "http://localhost:8191/v1"
	DefaultTimeout = 5 * time.Minute
)

var ErrSolverFault = errors.New("flare-solver")

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Result struct {
	Cookies   []Cookie
	UserAgent string
}

type solveRequest struct {
	CMD        string `json:"cmd"`
	URL        string `json:"url"`
	MaxTimeout int64  `json:"maxTimeout"`
}

type solveResponse struct {
	Status   string   `json:"status"`
	Message  string   `json:"message"`
	Solution solution `json:"solution"`
}

type solution struct {
	Cookies   []Cookie `json:"cookies"`
	UserAgent string   `json:"userAgent"`
}

type Solver struct {
	url     string
	timeout time.Duration
	client  *http.Client
}

func NewSolver() *Solver {
	return &Solver{
		url:     DefaultURL,
		timeout: DefaultTimeout,
		client:  &http.Client{Timeout: DefaultTimeout},
	}
}

func (s *Solver) GetCookies(ctx context.Context, targetURL string) (*Result, error) {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(solveRequest{
		CMD:        "request.get",
		URL:        targetURL,
		MaxTimeout: s.timeout.Milliseconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: request marshal: %v", ErrSolverFault, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, &body)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrSolverFault, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: http timeout", ErrSolverFault)
		}
		return nil, fmt.Errorf("%w: network call: %v", ErrSolverFault, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrSolverFault, resp.StatusCode)
	}

	var solveResp solveResponse
	if err := json.NewDecoder(resp.Body).Decode(&solveResp); err != nil {
		return nil, fmt.Errorf("%w: response decode: %v", ErrSolverFault, err)
	}

	if solveResp.Status != "ok" {
		return nil, fmt.Errorf("%w: target site failure: %s", ErrSolverFault, solveResp.Message)
	}

	return &Result{
		Cookies:   solveResp.Solution.Cookies,
		UserAgent: solveResp.Solution.UserAgent,
	}, nil
}
