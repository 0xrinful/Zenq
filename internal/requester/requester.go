package requester

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/0xrinful/Zenq/internal/requester/flare"
	"github.com/0xrinful/Zenq/internal/sources"
)

const defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:150.0) Gecko/20100101 Firefox/150.0"

type RequestMode int

const (
	ModeDefault RequestMode = iota
	ModeImage
)

type RequestOptions struct {
	Method  string
	Mode    RequestMode
	Headers map[string]string
	Body    []byte
}

type Requester struct {
	client     *http.Client
	solver     *flare.Solver
	config     sources.Config
	flareCache *flare.Result
	mu         sync.RWMutex
	group      singleflight.Group
}

func New(solver *flare.Solver, cfg sources.Config) *Requester {
	return &Requester{
		client: &http.Client{Timeout: 60 * time.Second},
		solver: solver,
		config: cfg,
	}
}

func (r *Requester) Get(ctx context.Context, url string) (*http.Response, error) {
	return r.Do(ctx, url, RequestOptions{Mode: ModeDefault})
}

func (r *Requester) GetImage(ctx context.Context, url string) (*http.Response, error) {
	return r.Do(ctx, url, RequestOptions{Mode: ModeImage})
}

func (r *Requester) Post(
	ctx context.Context,
	url string,
	contentType string,
	body []byte,
) (*http.Response, error) {
	return r.Do(ctx, url, RequestOptions{
		Method: http.MethodPost,
		Mode:   ModeDefault,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Body: body,
	})
}

func (r *Requester) Do(
	ctx context.Context,
	url string,
	opts RequestOptions,
) (*http.Response, error) {
	if r.config.NeedsFlare {
		if err := r.ensureSession(ctx, url); err != nil {
			return nil, err
		}
	}

	switch opts.Mode {
	case ModeImage:
		return r.doImage(ctx, url, opts)
	default:
		return r.doDefault(ctx, url, opts)
	}
}

func (r *Requester) doDefault(
	ctx context.Context,
	url string,
	opts RequestOptions,
) (*http.Response, error) {
	resp, err := r.send(ctx, url, opts)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		if err := r.refreshSession(ctx, url); err != nil {
			return nil, err
		}
		return r.send(ctx, url, opts)
	}

	return resp, nil
}

func (r *Requester) doImage(
	ctx context.Context,
	url string,
	opts RequestOptions,
) (*http.Response, error) {
	// retry up to 3 times with backoff before assuming flare block
	for attempt := range 3 {
		resp, err := r.send(ctx, url, opts)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusForbidden {
			return resp, nil
		}

		resp.Body.Close()

		if attempt == 2 {
			// 3rd 403 — real block, refresh flare
			if err := r.refreshSession(ctx, url); err != nil {
				return nil, err
			}
			return r.send(ctx, url, opts)
		}

		// temp 403 — small backoff and retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * 300 * time.Millisecond):
		}
	}

	return nil, fmt.Errorf("requester: image request failed after all retries: %s", url)
}

func (r *Requester) send(
	ctx context.Context,
	url string,
	opts RequestOptions,
) (*http.Response, error) {
	method := opts.Method
	if method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	if len(opts.Body) > 0 {
		body = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("requester: build request: %w", err)
	}

	// apply domain headers
	for k, v := range r.config.Headers {
		req.Header.Set(k, v)
	}
	// apply per-request headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	r.applySession(req)

	return r.client.Do(req)
}

func (r *Requester) ensureSession(ctx context.Context, url string) error {
	r.mu.RLock()
	cached := r.flareCache != nil
	r.mu.RUnlock()

	if cached {
		return nil
	}

	return r.refreshSession(ctx, url)
}

func (r *Requester) refreshSession(ctx context.Context, url string) error {
	_, err, _ := r.group.Do("flare-session", func() (any, error) {
		flareResult, err := r.solver.GetCookies(ctx, url)
		if err != nil {
			return nil, err
		}

		r.mu.Lock()
		r.flareCache = flareResult
		r.mu.Unlock()

		return nil, nil
	})

	return err
}

func (r *Requester) applySession(req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.flareCache == nil {
		req.Header.Set("User-Agent", defaultUserAgent)
		return
	}

	for _, c := range r.flareCache.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	req.Header.Set("User-Agent", r.flareCache.UserAgent)
}
