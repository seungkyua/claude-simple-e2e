// Package sdk 는 Gateway API와 통신하는 공통 HTTP 클라이언트를 제공한다.
// CLI, TUI, Gateway 모두에서 재사용 가능하도록 설계되었다.
package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// Client 는 Gateway API와 통신하는 HTTP 클라이언트이다
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	userAgent  string
	maxRetries int
}

// ClientOption 은 Client 생성 옵션이다
type ClientOption func(*Client)

// WithToken 은 인증 토큰을 설정한다
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithUserAgent 는 User-Agent 헤더를 설정한다
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) { c.userAgent = ua }
}

// WithMaxRetries 는 최대 재시도 횟수를 설정한다
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) { c.maxRetries = n }
}

// WithTimeout 은 요청 타임아웃을 설정한다
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// NewClient 는 새로운 SDK 클라이언트를 생성한다
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent:  "kcp-cli/1.0",
		maxRetries: 3,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetToken 은 인증 토큰을 업데이트한다
func (c *Client) SetToken(token string) {
	c.token = token
}

// doRequest 는 HTTP 요청을 실행하고 재시도 로직을 적용한다
// 재시도: 3회 + 지수 백오프 (1s, 2s, 4s)
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// 지수 백오프 대기
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			time.Sleep(backoff)
		}

		err := c.executeRequest(method, path, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// 재시도 불가능한 에러는 즉시 반환
		if !isRetryable(err) {
			return err
		}
	}

	return fmt.Errorf("최대 재시도 횟수 초과: %w", lastErr)
}

// executeRequest 는 단일 HTTP 요청을 실행한다
func (c *Client) executeRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("요청 직렬화 실패: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &RequestError{Err: err, Retryable: true}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("응답 읽기 실패: %w", err)
	}

	// 에러 응답 처리
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message,
				Detail:     errResp.Error.Detail,
				Retryable:  resp.StatusCode >= 500,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Code:       "UNKNOWN",
			Message:    string(respBody),
			Retryable:  resp.StatusCode >= 500,
		}
	}

	// 응답 파싱
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("응답 파싱 실패: %w", err)
		}
	}

	return nil
}

// Get 은 GET 요청을 실행한다
func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest(http.MethodGet, path, nil, result)
}

// Post 는 POST 요청을 실행한다
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.doRequest(http.MethodPost, path, body, result)
}

// Delete 는 DELETE 요청을 실행한다
func (c *Client) Delete(path string) error {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}

// RequestError 는 HTTP 요청 레벨 에러이다
type RequestError struct {
	Err       error
	Retryable bool
}

func (e *RequestError) Error() string { return e.Err.Error() }

// APIError 는 Gateway API 에러 응답이다
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Detail     string
	Retryable  bool
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s (%s)", e.StatusCode, e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
}

// isRetryable 은 에러가 재시도 가능한지 판단한다
func isRetryable(err error) bool {
	switch e := err.(type) {
	case *RequestError:
		return e.Retryable
	case *APIError:
		return e.Retryable
	default:
		return false
	}
}
