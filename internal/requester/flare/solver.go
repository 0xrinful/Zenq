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

var ErrChallengeFailed = errors.New("flare: challenge failed")

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

func New(url string) *Solver {
	return &Solver{
		url:     url,
		timeout: 180 * time.Second,
		client: &http.Client{
			Timeout: 180 * time.Second,
		},
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
		return nil, fmt.Errorf("%w: marshal request: %v", ErrChallengeFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, &body)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrChallengeFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: http call: %v", ErrChallengeFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: bad status %d", ErrChallengeFailed, resp.StatusCode)
	}

	var solveResp solveResponse
	if err := json.NewDecoder(resp.Body).Decode(&solveResp); err != nil {
		return nil, fmt.Errorf("%w: decode response: %v", ErrChallengeFailed, err)
	}

	if solveResp.Status != "ok" {
		return nil, fmt.Errorf("%w: %s", ErrChallengeFailed, solveResp.Message)
	}

	return &Result{
		Cookies:   solveResp.Solution.Cookies,
		UserAgent: solveResp.Solution.UserAgent,
	}, nil
}
