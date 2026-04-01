// 네트워크 관련 API 핸들러
// OpenStack SDK를 통해 Neutron API와 통신하여 네트워크, 서브넷, 라우터, 보안그룹 CRUD를 수행한다.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ossdk "github.com/kcp-cli/kcp-cli/pkg/sdk/openstack"
)

// NetworkHandler 는 네트워크, 서브넷, 라우터, 보안그룹 관련 API를 처리한다
type NetworkHandler struct {
	net *ossdk.NetworkService
}

// NewNetworkHandler 는 OpenStack SDK 클라이언트를 주입받아 NetworkHandler를 생성한다
func NewNetworkHandler(osClient *ossdk.Client) *NetworkHandler {
	return &NetworkHandler{net: ossdk.NewNetworkService(osClient)}
}

// ListNetworks 는 네트워크 목록을 조회한다 (Neutron /v2.0/networks)
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	items, err := h.net.ListNetworks()
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

// CreateNetwork 는 새로운 네트워크를 생성한다
func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.net.CreateNetwork(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteNetwork 는 지정된 네트워크를 삭제한다
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	id := c.Param("id")
	if err := h.net.DeleteNetwork(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListSubnets 는 서브넷 목록을 조회한다 (Neutron /v2.0/subnets)
func (h *NetworkHandler) ListSubnets(c *gin.Context) {
	items, err := h.net.ListSubnets()
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

// CreateSubnet 은 새로운 서브넷을 생성한다
func (h *NetworkHandler) CreateSubnet(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.net.CreateSubnet(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteSubnet 은 지정된 서브넷을 삭제한다
func (h *NetworkHandler) DeleteSubnet(c *gin.Context) {
	id := c.Param("id")
	if err := h.net.DeleteSubnet(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// ListRouters 는 라우터 목록을 조회한다 (Neutron /v2.0/routers)
func (h *NetworkHandler) ListRouters(c *gin.Context) {
	items, err := h.net.ListRouters()
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

// CreateRouter 는 새로운 라우터를 생성한다
func (h *NetworkHandler) CreateRouter(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.net.CreateRouter(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteRouter 는 지정된 라우터를 삭제한다
func (h *NetworkHandler) DeleteRouter(c *gin.Context) {
	id := c.Param("id")
	if err := h.net.DeleteRouter(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// AddRouterInterface 는 라우터에 서브넷 인터페이스를 추가한다
func (h *NetworkHandler) AddRouterInterface(c *gin.Context) {
	id := c.Param("id")

	var reqBody struct {
		SubnetID string `json:"subnet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "subnet_id 필드가 필요합니다: " + err.Error(), "status": 400},
		})
		return
	}

	// SDK의 AddRouterInterface 호출
	body := map[string]interface{}{
		"subnet_id": reqBody.SubnetID,
	}
	result, err := h.net.AddRouterInterface(id, body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// ListSecurityGroups 는 보안그룹 목록을 조회한다 (Neutron /v2.0/security-groups)
func (h *NetworkHandler) ListSecurityGroups(c *gin.Context) {
	items, err := h.net.ListSecurityGroups()
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

// CreateSecurityGroup 은 새로운 보안그룹을 생성한다
func (h *NetworkHandler) CreateSecurityGroup(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	result, err := h.net.CreateSecurityGroup(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}

// DeleteSecurityGroup 은 지정된 보안그룹을 삭제한다
func (h *NetworkHandler) DeleteSecurityGroup(c *gin.Context) {
	id := c.Param("id")
	if err := h.net.DeleteSecurityGroup(id); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// AddSecurityGroupRule 은 보안그룹에 규칙을 추가한다.
// URL의 :id를 security_group_id로 사용하여 요청에 포함시킨다.
func (h *NetworkHandler) AddSecurityGroupRule(c *gin.Context) {
	id := c.Param("id")

	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_REQUEST", "message": "요청 본문이 올바르지 않습니다: " + err.Error(), "status": 400},
		})
		return
	}

	// URL 파라미터의 보안그룹 ID를 요청 본문에 삽입
	reqBody["security_group_id"] = id

	result, err := h.net.AddSecurityGroupRule(reqBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "OPENSTACK_ERROR", "message": err.Error(), "status": 502},
		})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", result)
}
