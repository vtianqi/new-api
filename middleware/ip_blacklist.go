package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

var (
	ipBlacklist   = make(map[string]string) // ip -> 原因
	ipBlacklistMu sync.RWMutex
)

// AddIPToBlacklist 添加 IP 到黑名单
func AddIPToBlacklist(ip, reason string) {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()
	ipBlacklist[ip] = reason
	common.SysLog("IP黑名单：添加 " + ip + "，原因：" + reason)
}

// RemoveIPFromBlacklist 从黑名单移除 IP
func RemoveIPFromBlacklist(ip string) {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()
	delete(ipBlacklist, ip)
}

// GetBlacklist 获取黑名单列表
func GetBlacklist() map[string]string {
	ipBlacklistMu.RLock()
	defer ipBlacklistMu.RUnlock()
	result := make(map[string]string)
	for k, v := range ipBlacklist {
		result[k] = v
	}
	return result
}

// IPBlacklistMiddleware IP 黑名单中间件
func IPBlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		// 去掉端口
		if idx := strings.LastIndex(ip, ":"); idx != -1 && strings.Contains(ip, ".") {
			// IPv4 带端口，不处理（gin.ClientIP 一般不带端口）
		}

		ipBlacklistMu.RLock()
		reason, blocked := ipBlacklist[ip]
		ipBlacklistMu.RUnlock()

		if blocked {
			common.SysLog("IP黑名单拦截：" + ip + "，原因：" + reason)
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "您的IP已被限制访问，如有疑问请联系管理员",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
