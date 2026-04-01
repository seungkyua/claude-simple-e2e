// Package openstack 는 OpenStack API와 실제 통신하는 재사용 가능한 SDK 클라이언트를 제공한다.
// Keystone v3 인증을 통해 토큰을 발급받고, 각 서비스 엔드포인트를 조회한다.
package openstack

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultHTTPTimeout 은 HTTP 요청의 기본 타임아웃이다
	DefaultHTTPTimeout = 30 * time.Second
	// TokenRefreshThreshold 은 토큰 만료 전 자동 갱신 임계값이다
	TokenRefreshThreshold = 5 * time.Minute
)

// OSConfig 는 OpenStack 연결 설정이다 (openrc 호환)
type OSConfig struct {
	AuthURL         string // OS_AUTH_URL (Keystone v3 엔드포인트)
	AuthType        string // OS_AUTH_TYPE (password 등)
	Username        string // OS_USERNAME
	Password        string // OS_PASSWORD
	ProjectID       string // OS_PROJECT_ID
	ProjectName     string // OS_PROJECT_NAME
	ProjectDomainID string // OS_PROJECT_DOMAIN_ID
	UserDomainID    string // OS_USER_DOMAIN_ID
	DomainName      string // OS_USER_DOMAIN_NAME
	RegionName      string // OS_REGION_NAME
	Insecure        bool   // HTTPS 인증서 검증 무시 여부
}

// AuthToken 은 Keystone에서 발급받은 인증 토큰 정보이다
type AuthToken struct {
	Token     string
	ExpiresAt time.Time
	Catalog   []CatalogEntry
}

// CatalogEntry 는 서비스 카탈로그 항목이다
type CatalogEntry struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Endpoints []Endpoint `json:"endpoints"`
}

// Endpoint 는 서비스 엔드포인트이다
type Endpoint struct {
	URL       string `json:"url"`
	Interface string `json:"interface"`
	RegionID  string `json:"region_id"`
}

// Client 는 OpenStack API 클라이언트이다.
// Keystone 인증 토큰을 자동 관리하고, 서비스별 엔드포인트를 캐싱한다.
type Client struct {
	config     *OSConfig
	httpClient *http.Client
	auth       *AuthToken
	mu         sync.RWMutex
}

// NewClient 는 OpenStack 클라이언트를 생성한다.
// 초기 인증을 시도하되, 실패해도 클라이언트를 반환한다 (지연 인증 지원).
// 인증 실패 시 API 호출 시점에 재인증을 시도한다.
func NewClient(cfg *OSConfig) (*Client, error) {
	transport := &http.Transport{}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout:   DefaultHTTPTimeout,
			Transport: transport,
		},
	}

	// 초기 인증 시도 (실패해도 클라이언트는 반환)
	if err := c.Authenticate(); err != nil {
		return c, fmt.Errorf("초기 인증 실패 (API 호출 시 재시도합니다): %w", err)
	}

	return c, nil
}

// Authenticate 는 Keystone v3에서 토큰을 발급받는다
func (c *Client) Authenticate() error {
	// Keystone v3 인증 요청 본문 구성
	authReq := buildAuthRequest(c.config)

	body, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("인증 요청 직렬화 실패: %w", err)
	}

	authURL := normalizeAuthURL(c.config.AuthURL) + "/auth/tokens"
	req, err := http.NewRequest("POST", authURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("인증 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Keystone 연결 실패 (%s): %w", authURL, err)
	}
	defer resp.Body.Close()

	// 응답 본문을 먼저 한 번만 읽는다
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Keystone 응답 읽기 실패: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		// HTML 에러 응답인 경우 간결한 메시지로 변환
		errMsg := extractErrorMessage(respBody, resp.StatusCode)
		return fmt.Errorf("Keystone 인증 실패 (HTTP %d, URL: %s): %s", resp.StatusCode, authURL, errMsg)
	}

	// X-Subject-Token 헤더에서 토큰 추출
	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		return fmt.Errorf("Keystone 응답에 X-Subject-Token이 없습니다")
	}

	// 응답 본문에서 카탈로그, 만료 시간 파싱
	var tokenResp keystoneTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return fmt.Errorf("인증 응답 파싱 실패: %w", err)
	}

	expiresAt, _ := time.Parse(time.RFC3339, tokenResp.Token.ExpiresAt)

	c.mu.Lock()
	c.auth = &AuthToken{
		Token:     token,
		ExpiresAt: expiresAt,
		Catalog:   tokenResp.Token.Catalog,
	}
	c.mu.Unlock()

	return nil
}

// GetToken 은 유효한 인증 토큰을 반환한다. 만료 임박 시 자동 갱신한다.
func (c *Client) GetToken() (string, error) {
	c.mu.RLock()
	auth := c.auth
	c.mu.RUnlock()

	// 토큰이 없거나 만료 5분 전이면 재인증
	if auth == nil || time.Until(auth.ExpiresAt) < TokenRefreshThreshold {
		if err := c.Authenticate(); err != nil {
			return "", err
		}
		c.mu.RLock()
		auth = c.auth
		c.mu.RUnlock()
	}

	return auth.Token, nil
}

// GetEndpoint 는 서비스 타입에 해당하는 public 엔드포인트 URL을 반환한다
func (c *Client) GetEndpoint(serviceType string) (string, error) {
	c.mu.RLock()
	auth := c.auth
	c.mu.RUnlock()

	if auth == nil {
		return "", fmt.Errorf("인증되지 않은 상태입니다")
	}

	for _, entry := range auth.Catalog {
		if entry.Type == serviceType {
			for _, ep := range entry.Endpoints {
				if ep.Interface == "public" {
					// region 필터 (설정된 경우)
					if c.config.RegionName != "" && ep.RegionID != c.config.RegionName {
						continue
					}
					return ep.URL, nil
				}
			}
		}
	}

	return "", fmt.Errorf("서비스 '%s'의 엔드포인트를 찾을 수 없습니다", serviceType)
}

// DoRequest 는 OpenStack API에 인증된 HTTP 요청을 보낸다
func (c *Client) DoRequest(method, serviceType, path string, reqBody interface{}) ([]byte, int, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, 0, err
	}

	endpoint, err := c.GetEndpoint(serviceType)
	if err != nil {
		return nil, 0, err
	}

	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, 0, fmt.Errorf("요청 직렬화 실패: %w", err)
		}
		body = bytes.NewReader(data)
	}

	url := endpoint + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("OpenStack API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// --- Keystone 인증 요청/응답 구조체 ---

type keystoneAuthRequest struct {
	Auth keystoneAuth `json:"auth"`
}

type keystoneAuth struct {
	Identity keystoneIdentity `json:"identity"`
	Scope    *keystoneScope   `json:"scope,omitempty"`
}

type keystoneIdentity struct {
	Methods  []string          `json:"methods"`
	Password *keystonePassword `json:"password,omitempty"`
}

type keystonePassword struct {
	User keystoneUser `json:"user"`
}

type keystoneUser struct {
	Name     string          `json:"name"`
	Password string          `json:"password"`
	Domain   *keystoneDomain `json:"domain,omitempty"`
}

type keystoneDomain struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type keystoneScope struct {
	Project *keystoneProject `json:"project,omitempty"`
}

type keystoneProject struct {
	ID     string          `json:"id,omitempty"`
	Name   string          `json:"name,omitempty"`
	Domain *keystoneDomain `json:"domain,omitempty"`
}

type keystoneTokenResponse struct {
	Token struct {
		ExpiresAt string         `json:"expires_at"`
		Catalog   []CatalogEntry `json:"catalog"`
	} `json:"token"`
}

// buildAuthRequest 는 OSConfig에서 Keystone v3 인증 요청을 구성한다
func buildAuthRequest(cfg *OSConfig) *keystoneAuthRequest {
	req := &keystoneAuthRequest{
		Auth: keystoneAuth{
			Identity: keystoneIdentity{
				Methods: []string{"password"},
				Password: &keystonePassword{
					User: keystoneUser{
						Name:     cfg.Username,
						Password: cfg.Password,
					},
				},
			},
		},
	}

	// 사용자 도메인 설정 (OS_USER_DOMAIN_ID)
	if cfg.UserDomainID != "" {
		req.Auth.Identity.Password.User.Domain = &keystoneDomain{ID: cfg.UserDomainID}
	} else if cfg.DomainName != "" {
		req.Auth.Identity.Password.User.Domain = &keystoneDomain{Name: cfg.DomainName}
	}

	// 프로젝트 스코프 설정
	if cfg.ProjectID != "" {
		proj := &keystoneProject{ID: cfg.ProjectID}
		// 프로젝트 도메인 설정 (OS_PROJECT_DOMAIN_ID)
		if cfg.ProjectDomainID != "" {
			proj.Domain = &keystoneDomain{ID: cfg.ProjectDomainID}
		}
		req.Auth.Scope = &keystoneScope{Project: proj}
	} else if cfg.ProjectName != "" {
		proj := &keystoneProject{Name: cfg.ProjectName}
		if cfg.ProjectDomainID != "" {
			proj.Domain = &keystoneDomain{ID: cfg.ProjectDomainID}
		}
		req.Auth.Scope = &keystoneScope{Project: proj}
	}

	return req
}

// extractErrorMessage 는 Keystone 에러 응답에서 의미 있는 메시지를 추출한다.
// HTML 응답인 경우 간결한 메시지로 변환한다.
func extractErrorMessage(body []byte, statusCode int) string {
	// JSON 에러 응답 시도
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
		return errResp.Error.Message
	}

	// HTML 응답이거나 파싱 불가한 경우 상태 코드 기반 메시지
	switch statusCode {
	case 401:
		return "인증 정보가 올바르지 않습니다 (username/password 확인)"
	case 403:
		return "접근이 거부되었습니다 (권한 확인)"
	case 404:
		return "Keystone 엔드포인트를 찾을 수 없습니다 (auth_url 확인)"
	case 500:
		return "OpenStack 서버 내부 오류 (Keystone 서비스 상태 확인 필요)"
	case 503:
		return "OpenStack 서비스를 사용할 수 없습니다 (서비스 상태 확인)"
	default:
		// 200자 이내로 잘라서 반환
		s := string(body)
		if len(s) > 200 {
			s = s[:200] + "..."
		}
		return s
	}
}

// normalizeAuthURL 은 Keystone Auth URL을 정규화한다.
// /v3가 포함되어 있지 않으면 자동으로 추가한다.
func normalizeAuthURL(authURL string) string {
	url := strings.TrimRight(authURL, "/")
	if strings.HasSuffix(url, "/v3") {
		return url
	}
	return url + "/v3"
}

// checkStatusError 는 HTTP 상태 코드가 에러 범위(>= 400)인 경우 에러를 반환한다
func checkStatusError(data []byte, statusCode int, operation string) error {
	if statusCode >= 400 {
		errMsg := extractErrorMessage(data, statusCode)
		return fmt.Errorf("%s 실패 (HTTP %d): %s", operation, statusCode, errMsg)
	}
	return nil
}
