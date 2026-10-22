package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

var HttpCl *HttpClient

// perRequestTimeout bounds a single RawRequest attempt. It is deliberately
// short because these calls run inside the per-action macro loop: a silent
// stall on the remote helper must fail fast and visibly rather than freezing
// the loop for the full client timeout.
const perRequestTimeout = 3 * time.Second

type HttpClient struct {
	Client  *http.Client
	baseUrl string
}

func (cl *HttpClient) RawRequest(path string, method string, body io.Reader) (result []byte, err error) {
	const maxRetries = 1
	var resp *http.Response
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), perRequestTimeout)
		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, method, cl.baseUrl+path, body)
		if err != nil {
			cancel()
			return nil, err
		}
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err = cl.Client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			res, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			cancel()
			if readErr != nil {
				return nil, readErr
			}
			return res, nil
		}

		// Collect a diagnostic message. We intentionally avoid writing to
		// stdout here: during a stall only the error returned to the caller
		// (which the callers log via zap) should surface in app.log.
		if err != nil {
			err = fmt.Errorf("attempt %d: request failed: %w", attempt, err)
		} else {
			status := resp.Status
			resp.Body.Close()
			err = fmt.Errorf("attempt %d: unexpected status: %s", attempt, status)
		}
		cancel()

		time.Sleep(time.Second / 2)
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}

func IniHttpClient(baseUrl string) {
	HttpCl = &HttpClient{
		baseUrl: baseUrl,
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second, // Persistent connections
				}).DialContext,
				MaxIdleConns:        10,               // Total idle connections
				IdleConnTimeout:     90 * time.Second, // Keep idle connection alive
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
}
