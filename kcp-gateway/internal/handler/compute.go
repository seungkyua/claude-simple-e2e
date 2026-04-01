// 컴퓨트(서버/인스턴스) 관련 API 핸들러
// OpenStack SDK를 통해 Nova API와 통신하여 서버, 플레이버 CRUD를 수행한다.
// 서버 목록/상세 응답에 flavor_name, image_name을 자동으로 enrichment한다.
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-gateway/pkg/openstack"
)

// ComputeHandler 는 컴퓨트(서버, 플레이버) 관련 API를 처리한다
type ComputeHandler struct {
	compute *ossdk.ComputeService
	image   *ossdk.ImageService
	net     *ossdk.NetworkService
}

// NewComputeHandler 는 OpenStack SDK 클라이언트를 주입받아 ComputeHandler를 생성한다
func NewComputeHandler(osClient *ossdk.Client) *ComputeHandler {
	return &ComputeHandler{
		compute: ossdk.NewComputeService(osClient),
		image:   ossdk.NewImageService(osClient),
		net:     ossdk.NewNetworkService(osClient),
	}
}

// fetchFlavorNameMap 은 Flavor ID→Name 매핑을 구축한다
func (h *ComputeHandler) fetchFlavorNameMap() map[string]string {
	m := make(map[string]string)
	items, err := h.compute.ListFlavors()
	if err != nil {
		return m
	}
	for _, raw := range items {
		var item map[string]interface{}
		if json.Unmarshal(raw, &item) != nil {
			continue
		}
		id, _ := item["id"].(string)
		name, _ := item["name"].(string)
		if id != "" && name != "" {
			m[id] = name
		}
	}
	return m
}

// fetchImageNameMap 은 Image ID→Name 매핑을 구축한다
func (h *ComputeHandler) fetchImageNameMap() map[string]string {
	m := make(map[string]string)
	items, err := h.image.ListImages()
	if err != nil {
		return m
	}
	for _, raw := range items {
		var item map[string]interface{}
		if json.Unmarshal(raw, &item) != nil {
			continue
		}
		id, _ := item["id"].(string)
		name, _ := item["name"].(string)
		if id != "" && name != "" {
			m[id] = name
		}
	}
	return m
}

// enrichServers 는 서버 목록에 flavor_name, image_name을 추가한다
func (h *ComputeHandler) enrichServers(items []json.RawMessage) []json.RawMessage {
	// flavor, image 이름 매핑을 병렬로 조회
	var flavorMap, imageMap map[string]string
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		flavorMap = h.fetchFlavorNameMap()
	}()
	go func() {
		defer wg.Done()
		imageMap = h.fetchImageNameMap()
	}()
	wg.Wait()

	enriched := make([]json.RawMessage, 0, len(items))
	for _, raw := range items {
		var server map[string]interface{}
		if json.Unmarshal(raw, &server) != nil {
			enriched = append(enriched, raw)
			continue
		}

		// flavor enrichment
		if flavor, ok := server["flavor"].(map[string]interface{}); ok {
			if fid, ok := flavor["id"].(string); ok {
				if name, ok := flavorMap[fid]; ok {
					flavor["name"] = name
				}
			}
		}

		// image enrichment
		if image, ok := server["image"].(map[string]interface{}); ok {
			if iid, ok := image["id"].(string); ok {
				if name, ok := imageMap[iid]; ok {
					image["name"] = name
				}
			}
		}

		// addresses를 읽기 쉬운 networks 문자열로 변환
		if addresses, ok := server["addresses"].(map[string]interface{}); ok {
			var parts []string
			for netName, addrsRaw := range addresses {
				addrs, ok := addrsRaw.([]interface{})
				if !ok {
					continue
				}
				var addrStrs []string
				for _, a := range addrs {
					if am, ok := a.(map[string]interface{}); ok {
						if addr, ok := am["addr"].(string); ok {
							addrStrs = append(addrStrs, addr)
						}
					}
				}
				parts = append(parts, fmt.Sprintf("%s=%s", netName, strings.Join(addrStrs, ", ")))
			}
			server["networks"] = strings.Join(parts, "; ")
		}

		data, err := json.Marshal(server)
		if err != nil {
			enriched = append(enriched, raw)
			continue
		}
		enriched = append(enriched, data)
	}
	return enriched
}

// ListServers 는 서버 목록을 조회한다 (Nova /servers/detail)
// Gateway에서 flavor_name, image_name, networks를 자동 enrichment한다
func (h *ComputeHandler) ListServers(c *gin.Context) {
	items, err := h.compute.ListServers()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}

	// enrichment 적용
	enriched := h.enrichServers(items)

	c.JSON(http.StatusOK, kcpListResponse{
		Items: enriched,
		Pagination: kcpPagination{
			Page:  1,
			Size:  len(enriched),
			Total: len(enriched),
		},
	})
}

// GetServer 는 특정 서버의 상세 정보를 조회한다 (enrichment 포함)
func (h *ComputeHandler) GetServer(c *gin.Context) {
	id := c.Param("id")
	serverRaw, err := h.compute.GetServer(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}

	// 단일 서버 enrichment
	enriched := h.enrichServers([]json.RawMessage{serverRaw})
	if len(enriched) > 0 {
		c.Data(http.StatusOK, "application/json; charset=utf-8", enriched[0])
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", serverRaw)
}

// createServerRequest 는 KCP 서버 생성 요청 형식이다
type createServerRequest struct {
	Name             string   `json:"name" binding:"required"`
	FlavorID         string   `json:"flavorId" binding:"required"`
	ImageID          string   `json:"imageId" binding:"required"`
	NetworkIDs       []string `json:"networkIds,omitempty"`
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`
	KeyName          string   `json:"keyName,omitempty"`
}

// isUUID 는 문자열이 UUID 형식��지 간단히 판별한다
func isUUID(s string) bool {
	// UUID: 8-4-4-4-12 또는 32자 hex
	if len(s) == 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-' {
		return true
	}
	return false
}

// resolveNameToID 는 리소스 목록에서 이름으로 ID를 찾는다.
// 이미 UUID 형식이면 그대로 반환한다.
func resolveNameToID(nameOrID string, items []json.RawMessage) string {
	if isUUID(nameOrID) {
		return nameOrID
	}
	for _, raw := range items {
		var item map[string]interface{}
		if json.Unmarshal(raw, &item) != nil {
			continue
		}
		name, _ := item["name"].(string)
		id, _ := item["id"].(string)
		if name == nameOrID {
			return id
		}
	}
	return nameOrID // 찾지 못하면 원본 반환
}

// CreateServer 는 새로운 서버를 생성한다.
// 이름 또는 UUID를 모두 지원한다 (flavor, image, network, security-group).
// KCP 요청 형식을 Nova API 형식으로 변환한 뒤 호출한다.
func (h *ComputeHandler) CreateServer(c *gin.Context) {
	var req createServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	// flavor 이름 → ID 변환
	flavorRef := req.FlavorID
	if !isUUID(flavorRef) {
		if flavors, err := h.compute.ListFlavors(); err == nil {
			flavorRef = resolveNameToID(flavorRef, flavors)
		}
	}

	// image 이름 → ID 변환
	imageRef := req.ImageID
	if !isUUID(imageRef) {
		if images, err := h.image.ListImages(); err == nil {
			imageRef = resolveNameToID(imageRef, images)
		}
	}

	// Nova API 형식으로 변환
	novaServer := map[string]interface{}{
		"name":      req.Name,
		"flavorRef": flavorRef,
		"imageRef":  imageRef,
	}

	// 네트워크 이름 → UUID 변환
	if len(req.NetworkIDs) > 0 {
		var networkList []json.RawMessage
		needResolve := false
		for _, nid := range req.NetworkIDs {
			if !isUUID(nid) {
				needResolve = true
				break
			}
		}
		if needResolve {
			networkList, _ = h.net.ListNetworks()
		}

		var networks []map[string]string
		for _, nid := range req.NetworkIDs {
			resolved := nid
			if !isUUID(nid) && networkList != nil {
				resolved = resolveNameToID(nid, networkList)
			}
			networks = append(networks, map[string]string{"uuid": resolved})
		}
		novaServer["networks"] = networks
	}

	// SSH 키 설정
	if req.KeyName != "" {
		novaServer["key_name"] = req.KeyName
	}

	// 보안그룹 설정 (이름 또는 UUID — Nova는 name으로 받음)
	if len(req.SecurityGroupIDs) > 0 {
		var sgs []map[string]string
		for _, sg := range req.SecurityGroupIDs {
			sgs = append(sgs, map[string]string{"name": sg})
		}
		novaServer["security_groups"] = sgs
	}

	createResult, err := h.compute.CreateServer(novaServer)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}

	// 생성 응답에서 서버 ID 추출
	var created map[string]interface{}
	if json.Unmarshal(createResult, &created) != nil {
		c.Data(http.StatusCreated, "application/json; charset=utf-8", createResult)
		return
	}

	serverID, _ := created["id"].(string)
	if serverID == "" {
		c.Data(http.StatusCreated, "application/json; charset=utf-8", createResult)
		return
	}

	// 생성 직후 상세 조회 (OpenStack CLI와 동일한 동작)
	// adminPass는 생성 응답에만 포함되므로 보존한다
	adminPass, _ := created["adminPass"].(string)

	detail, err := h.compute.GetServer(serverID)
	if err != nil {
		// 상세 조회 실패 시 생성 응답 그대로 반환
		c.Data(http.StatusCreated, "application/json; charset=utf-8", createResult)
		return
	}

	// 상세 응답에 adminPass 추가
	var detailMap map[string]interface{}
	if json.Unmarshal(detail, &detailMap) == nil && adminPass != "" {
		detailMap["adminPass"] = adminPass
		if merged, err := json.Marshal(detailMap); err == nil {
			detail = merged
		}
	}

	// enrichment 적용 후 응답
	enriched := h.enrichServers([]json.RawMessage{detail})
	if len(enriched) > 0 {
		c.Data(http.StatusCreated, "application/json; charset=utf-8", enriched[0])
		return
	}
	c.Data(http.StatusCreated, "application/json; charset=utf-8", detail)
}

// DeleteServer 는 지정된 서버를 삭제한다
func (h *ComputeHandler) DeleteServer(c *gin.Context) {
	id := c.Param("id")
	if err := h.compute.DeleteServer(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// ServerAction 은 서버에 대한 액션(시작, 중지, 재부팅 등)을 수행한다
func (h *ComputeHandler) ServerAction(c *gin.Context) {
	id := c.Param("id")

	var reqBody struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "action 필드가 필요합니다: " + err.Error(), "status": 400},
		})
		return
	}

	var novaAction map[string]interface{}
	switch reqBody.Action {
	case "start":
		novaAction = map[string]interface{}{"os-start": nil}
	case "stop":
		novaAction = map[string]interface{}{"os-stop": nil}
	case "reboot":
		novaAction = map[string]interface{}{
			"reboot": map[string]string{"type": "SOFT"},
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_ACTION",
				"message": "지원하지 않는 액션입니다. start, stop, reboot만 가능합니다",
				"status":  400,
			},
		})
		return
	}

	if err := h.compute.ServerAction(id, novaAction); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusAccepted)
}

// ListFlavors 는 플레이버(인스턴스 사양) 목록을 조회한다
func (h *ComputeHandler) ListFlavors(c *gin.Context) {
	items, err := h.compute.ListFlavors()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}

	c.JSON(http.StatusOK, kcpListResponse{
		Items: items,
		Pagination: kcpPagination{
			Page:  1,
			Size:  len(items),
			Total: len(items),
		},
	})
}

// CreateFlavor 는 새로운 플레이버를 생성한다
func (h *ComputeHandler) CreateFlavor(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.compute.CreateFlavor(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteFlavor 는 지정된 플레이버를 삭제한다
func (h *ComputeHandler) DeleteFlavor(c *gin.Context) {
	id := c.Param("id")
	if err := h.compute.DeleteFlavor(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}
