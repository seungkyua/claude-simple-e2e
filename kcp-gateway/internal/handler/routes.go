// Package handler 는 Gateway API 핸들러를 제공한다.
package handler

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/kcp-cli/kcp-gateway/config"
	"github.com/kcp-cli/kcp-gateway/internal/openstack"
)

// RegisterAuthRoutes 는 인증 관련 라우트를 등록한다
func RegisterAuthRoutes(rg *gin.RouterGroup, db *sql.DB, cfg *config.Config) {
	auth := rg.Group("/auth")
	h := NewAuthHandler(db, cfg)
	auth.POST("/login", h.Login)
	auth.POST("/logout", h.Logout)
	auth.POST("/refresh", h.Refresh)
}

// RegisterComputeRoutes 는 Compute 관련 라우트를 등록한다
func RegisterComputeRoutes(rg *gin.RouterGroup, osClient *openstack.Client) {
	compute := rg.Group("/compute")
	h := NewComputeHandler(osClient)
	compute.GET("/servers", h.ListServers)
	compute.GET("/servers/:id", h.GetServer)
	compute.POST("/servers", h.CreateServer)
	compute.DELETE("/servers/:id", h.DeleteServer)
	compute.POST("/servers/:id/action", h.ServerAction)
	compute.GET("/flavors", h.ListFlavors)
	compute.POST("/flavors", h.CreateFlavor)
	compute.DELETE("/flavors/:id", h.DeleteFlavor)
}

// RegisterNetworkRoutes 는 Network 관련 라우트를 등록한다
func RegisterNetworkRoutes(rg *gin.RouterGroup, osClient *openstack.Client) {
	net := rg.Group("/network")
	h := NewNetworkHandler(osClient)
	net.GET("/networks", h.ListNetworks)
	net.POST("/networks", h.CreateNetwork)
	net.DELETE("/networks/:id", h.DeleteNetwork)
	net.GET("/subnets", h.ListSubnets)
	net.POST("/subnets", h.CreateSubnet)
	net.DELETE("/subnets/:id", h.DeleteSubnet)
	net.GET("/routers", h.ListRouters)
	net.POST("/routers", h.CreateRouter)
	net.DELETE("/routers/:id", h.DeleteRouter)
	net.POST("/routers/:id/add-interface", h.AddRouterInterface)
	net.GET("/security-groups", h.ListSecurityGroups)
	net.POST("/security-groups", h.CreateSecurityGroup)
	net.DELETE("/security-groups/:id", h.DeleteSecurityGroup)
	net.POST("/security-groups/:id/rules", h.AddSecurityGroupRule)
}

// RegisterStorageRoutes 는 Storage 관련 라우트를 등록한다
func RegisterStorageRoutes(rg *gin.RouterGroup, osClient *openstack.Client) {
	storage := rg.Group("/storage")
	h := NewStorageHandler(osClient)
	storage.GET("/volumes", h.ListVolumes)
	storage.POST("/volumes", h.CreateVolume)
	storage.DELETE("/volumes/:id", h.DeleteVolume)
	storage.POST("/volumes/:id/attach", h.AttachVolume)
	storage.POST("/volumes/:id/detach", h.DetachVolume)
	storage.GET("/snapshots", h.ListSnapshots)
	storage.POST("/snapshots", h.CreateSnapshot)
	storage.DELETE("/snapshots/:id", h.DeleteSnapshot)
}

// RegisterIdentityRoutes 는 Identity 관련 라우트를 등록한다
func RegisterIdentityRoutes(rg *gin.RouterGroup, osClient *openstack.Client) {
	identity := rg.Group("/identity")
	h := NewIdentityHandler(osClient)
	identity.GET("/projects", h.ListProjects)
	identity.POST("/projects", h.CreateProject)
	identity.DELETE("/projects/:id", h.DeleteProject)
	identity.GET("/users", h.ListUsers)
	identity.POST("/users", h.CreateUser)
	identity.DELETE("/users/:id", h.DeleteUser)
	identity.POST("/roles/assign", h.AssignRole)
	identity.DELETE("/roles/revoke", h.RevokeRole)
}

// RegisterImageRoutes 는 Image 관련 라우트를 등록한다
func RegisterImageRoutes(rg *gin.RouterGroup, osClient *openstack.Client) {
	image := rg.Group("/image")
	h := NewImageHandler(osClient)
	image.GET("/images", h.ListImages)
	image.GET("/images/:id", h.GetImage)
	image.POST("/images", h.UploadImage)
	image.DELETE("/images/:id", h.DeleteImage)
}

// RegisterAuditRoutes 는 감사 로그 라우트를 등록한다
func RegisterAuditRoutes(rg *gin.RouterGroup, db *sql.DB) {
	audit := rg.Group("/audit")
	h := NewAuditHandler(db)
	audit.GET("/logs", h.ListLogs)
}

// RegisterStatsRoutes 는 통계 라우트를 등록한다
func RegisterStatsRoutes(rg *gin.RouterGroup, osClient *openstack.Client, db *sql.DB) {
	stats := rg.Group("/stats")
	h := NewStatsHandler(osClient, db)
	stats.GET("/dashboard", h.Dashboard)
}
