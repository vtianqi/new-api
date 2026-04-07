package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

// GetIPBlacklist 获取黑名单列表
func GetIPBlacklist(c *gin.Context) {
	list := middleware.GetBlacklist()
	items := make([]gin.H, 0, len(list))
	for ip, reason := range list {
		items = append(items, gin.H{"ip": ip, "reason": reason})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": items})
}

// AddIPBlacklist 添加 IP 到黑名单
func AddIPBlacklist(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}
	if req.Reason == "" {
		req.Reason = "管理员手动封禁"
	}
	middleware.AddIPToBlacklist(req.IP, req.Reason)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已添加到黑名单"})
}

// RemoveIPBlacklist 从黑名单移除 IP
func RemoveIPBlacklist(c *gin.Context) {
	ip := c.Param("ip")
	middleware.RemoveIPFromBlacklist(ip)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已从黑名单移除"})
}
