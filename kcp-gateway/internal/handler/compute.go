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
	ossdk "github.com/kcp-cli/kcp-cli/pkg/sdk/openstack"
)

// ComputeHandler 는 컴퓨트(서버, 플레이버) 관련 API를 처리한다
type ComputeHandler struct {
	compute *ossdk.ComputeService
	image   *ossdk.ImageService
}

// NewComputeHandler 는 OpenStack SDK 클라이언트를 주입받아 ComputeHandler를 생성한다
func NewComputeHandler(osClient *ossdk.Client) *ComputeHandler {
	return &ComputeHandler{
		compute: ossdk.NewComputeService(osClient),
		image:   ossdk.NewImageService(osClient),
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

// CreateServer 는 새로운 서버를 생성한다 (Nova 서버 생성 요청)
func (h *ComputeHandler) CreateServer(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.compute.CreateServer(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
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
