// 네트워크 관련 API 핸들러
// OpenStack Neutron API와 실제 통신하여 네트워크, 서브넷, 라우터, 보안그룹 CRUD를 수행한다.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// NetworkHandler 는 네트워크, 서브넷, 라우터, 보안그룹 관련 API를 처리한다
type NetworkHandler struct {
	osClient *openstack.Client
}

// NewNetworkHandler 는 OpenStack 클라이언트를 주입받아 NetworkHandler를 생성한다
func NewNetworkHandler(osClient *openstack.Client) *NetworkHandler {
	return &NetworkHandler{osClient: osClient}
}

// ListNetworks 는 네트워크 목록을 조회한다 (Neutron /v2.0/networks)
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	data, status, err := h.osClient.DoRequest("GET", "network", "/v2.0/networks", nil)
	forwardOSListResponse(c, data, status, err, "networks")
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

	// Neutron API 형식으로 래핑: {"network": {...}}
	neutronReq := map[string]interface{}{
		"network": reqBody,
	}

	data, status, err := h.osClient.DoRequest("POST", "network", "/v2.0/networks", neutronReq)
	forwardOSSingleResponse(c, data, status, err, "network")
}

// DeleteNetwork 는 지정된 네트워크를 삭제한다
func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.osClient.DoRequest("DELETE", "network", "/v2.0/networks/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// ListSubnets 는 서브넷 목록을 조회한다 (Neutron /v2.0/subnets)
func (h *NetworkHandler) ListSubnets(c *gin.Context) {
	data, status, err := h.osClient.DoRequest("GET", "network", "/v2.0/subnets", nil)
	forwardOSListResponse(c, data, status, err, "subnets")
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

	// Neutron API 형식으로 래핑: {"subnet": {...}}
	neutronReq := map[string]interface{}{
		"subnet": reqBody,
	}

	data, status, err := h.osClient.DoRequest("POST", "network", "/v2.0/subnets", neutronReq)
	forwardOSSingleResponse(c, data, status, err, "subnet")
}

// DeleteSubnet 은 지정된 서브넷을 삭제한다
func (h *NetworkHandler) DeleteSubnet(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.osClient.DoRequest("DELETE", "network", "/v2.0/subnets/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// ListRouters 는 라우터 목록을 조회한다 (Neutron /v2.0/routers)
func (h *NetworkHandler) ListRouters(c *gin.Context) {
	data, status, err := h.osClient.DoRequest("GET", "network", "/v2.0/routers", nil)
	forwardOSListResponse(c, data, status, err, "routers")
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

	// Neutron API 형식으로 래핑: {"router": {...}}
	neutronReq := map[string]interface{}{
		"router": reqBody,
	}

	data, status, err := h.osClient.DoRequest("POST", "network", "/v2.0/routers", neutronReq)
	forwardOSSingleResponse(c, data, status, err, "router")
}

// DeleteRouter 는 지정된 라우터를 삭제한다
func (h *NetworkHandler) DeleteRouter(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.osClient.DoRequest("DELETE", "network", "/v2.0/routers/"+id, nil)
	forwardOSResponse(c, data, status, err)
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

	// Neutron 라우터 인터페이스 추가 요청
	neutronReq := map[string]interface{}{
		"subnet_id": reqBody.SubnetID,
	}

	data, status, err := h.osClient.DoRequest("PUT", "network", "/v2.0/routers/"+id+"/add_router_interface", neutronReq)
	forwardOSResponse(c, data, status, err)
}

// ListSecurityGroups 는 보안그룹 목록을 조회한다 (Neutron /v2.0/security-groups)
func (h *NetworkHandler) ListSecurityGroups(c *gin.Context) {
	data, status, err := h.osClient.DoRequest("GET", "network", "/v2.0/security-groups", nil)
	forwardOSListResponse(c, data, status, err, "security_groups")
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

	// Neutron API 형식으로 래핑: {"security_group": {...}}
	neutronReq := map[string]interface{}{
		"security_group": reqBody,
	}

	data, status, err := h.osClient.DoRequest("POST", "network", "/v2.0/security-groups", neutronReq)
	forwardOSSingleResponse(c, data, status, err, "security_group")
}

// DeleteSecurityGroup 은 지정된 보안그룹을 삭제한다
func (h *NetworkHandler) DeleteSecurityGroup(c *gin.Context) {
	id := c.Param("id")
	data, status, err := h.osClient.DoRequest("DELETE", "network", "/v2.0/security-groups/"+id, nil)
	forwardOSResponse(c, data, status, err)
}

// AddSecurityGroupRule 은 보안그룹에 규칙을 추가한다.
// URL의 :id를 security_group_id로 사용하여 Neutron 요청에 포함시킨다.
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

	// Neutron API 형식으로 래핑: {"security_group_rule": {...}}
	neutronReq := map[string]interface{}{
		"security_group_rule": reqBody,
	}

	data, status, err := h.osClient.DoRequest("POST", "network", "/v2.0/security-group-rules", neutronReq)
	forwardOSSingleResponse(c, data, status, err, "security_group_rule")
}
