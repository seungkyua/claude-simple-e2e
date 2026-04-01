package sdk

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNetworkListNetworks 는 네트워크 목록 조회 응답 파싱을 검증한다
func TestNetworkListNetworks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/network/networks" {
			t.Errorf("요청 경로가 '/network/networks'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("요청 메서드가 GET이어야 하지만 '%s'이다", r.Method)
		}

		resp := ListResponse[Network]{
			Items: []Network{
				{ID: "net-001", Name: "public-net", Status: "ACTIVE", AdminStateUp: true},
				{ID: "net-002", Name: "private-net", Status: "ACTIVE", Shared: false},
			},
			Pagination: Pagination{Page: 1, Size: 10, Total: 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	nc := NewNetworkClient(client)

	result, err := nc.ListNetworks(nil)
	if err != nil {
		t.Fatalf("ListNetworks 실패: %v", err)
	}

	if len(result.Items) != 2 {
		t.Fatalf("네트워크 수가 2개여야 하지만 %d개이다", len(result.Items))
	}
	if result.Items[0].Name != "public-net" {
		t.Errorf("첫 번째 네트워크 이름이 'public-net'이어야 하지만 '%s'이다", result.Items[0].Name)
	}
	if !result.Items[0].AdminStateUp {
		t.Error("첫 번째 네트워크의 AdminStateUp이 true이어야 한다")
	}
}

// TestNetworkCreateSubnet 은 서브넷 생성 요청을 검증한다
func TestNetworkCreateSubnet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/network/subnets" {
			t.Errorf("요청 경로가 '/network/subnets'이어야 하지만 '%s'이다", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("요청 메서드가 POST이어야 하지만 '%s'이다", r.Method)
		}

		// 요청 본문 검증
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("요청 본문 읽기 실패: %v", err)
		}

		var req CreateSubnetRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("요청 본문 파싱 실패: %v", err)
		}
		if req.Name != "my-subnet" {
			t.Errorf("서브넷 이름이 'my-subnet'이어야 하지만 '%s'이다", req.Name)
		}
		if req.CIDR != "10.0.0.0/24" {
			t.Errorf("CIDR이 '10.0.0.0/24'이어야 하지만 '%s'이다", req.CIDR)
		}

		resp := Subnet{
			ID:        "sub-001",
			Name:      req.Name,
			NetworkID: req.NetworkID,
			CIDR:      req.CIDR,
			IPVersion: req.IPVersion,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, WithMaxRetries(0))
	nc := NewNetworkClient(client)

	subnet, err := nc.CreateSubnet(&CreateSubnetRequest{
		Name:      "my-subnet",
		NetworkID: "net-001",
		CIDR:      "10.0.0.0/24",
		IPVersion: 4,
	})
	if err != nil {
		t.Fatalf("CreateSubnet 실패: %v", err)
	}
	if subnet.ID != "sub-001" {
		t.Errorf("생성된 서브넷 ID가 'sub-001'이어야 하지만 '%s'이다", subnet.ID)
	}
	if subnet.CIDR != "10.0.0.0/24" {
		t.Errorf("생성된 서브넷 CIDR이 '10.0.0.0/24'이어야 하지만 '%s'이다", subnet.CIDR)
	}
}
