package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// CheckUserDailyQuotaAlert 检查用户今日用量是否超过阈值，超过则发邮件告警
func CheckUserDailyQuotaAlert(userId int, username string, todayQuota int64) {
	const alertThreshold = 5000000
	if todayQuota < alertThreshold {
		return
	}

	cacheKey := fmt.Sprintf("quota_alert:%d:%s", userId, time.Now().Format("2006-01-02"))
	if common.RedisEnabled {
		ctx := context.Background()
		exists, _ := common.RDB.Exists(ctx, cacheKey).Result()
		if exists > 0 {
			return
		}
		common.RDB.Set(ctx, cacheKey, 1, 24*time.Hour)
	}

	rootUser, err := model.GetUserById(1, false)
	if err != nil || rootUser == nil || rootUser.Email == "" {
		return
	}
	adminEmail := rootUser.Email

	yuan := float64(todayQuota) / common.QuotaPerUnit
	subject := fmt.Sprintf("【用量告警】用户 %s 今日消耗 %.2f 元", username, yuan)
	content := fmt.Sprintf(`
<h3>用量告警通知</h3>
<p>用户 <strong>%s</strong> (ID: %d) 今日用量已超过告警阈值：</p>
<ul>
  <li>今日消耗 Quota：%d</li>
  <li>折合金额：约 %.4f 元</li>
  <li>告警时间：%s</li>
</ul>
<p>请登录管理后台查看详情。</p>
`, username, userId, todayQuota, yuan, time.Now().Format("2006-01-02 15:04:05"))

	go func() {
		if err := common.SendEmail(subject, adminEmail, content); err != nil {
			common.SysError("发送用量告警邮件失败: " + err.Error())
		}
	}()
}

// GetQuotaAlertConfig 获取告警配置
func GetQuotaAlertConfig(c *gin.Context) {
	rootUser, _ := model.GetUserById(1, false)
	adminEmail := ""
	if rootUser != nil {
		adminEmail = rootUser.Email
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"alert_threshold": 5000000,
			"alert_enabled":   adminEmail != "",
			"admin_email":     adminEmail,
		},
	})
}

// GetUserTodayQuota 查询某用户今日用量（管理员用）
func GetUserTodayQuota(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "缺少username参数"})
		return
	}

	today := time.Now().Format("2006-01-02")
	startTime := time.Now().Truncate(24 * time.Hour).Unix()

	var result struct {
		Quota int64 `gorm:"column:quota"`
		Calls int64 `gorm:"column:calls"`
	}
	model.LOG_DB.Table("logs").
		Select("sum(quota) as quota, count(*) as calls").
		Where("username = ? AND created_at >= ? AND type = 2", username, startTime).
		Scan(&result)

	yuan := float64(result.Quota) / common.QuotaPerUnit

	userId, _ := strconv.Atoi(c.Query("user_id"))
	if userId > 0 {
		go CheckUserDailyQuotaAlert(userId, username, result.Quota)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"date":  today,
			"quota": result.Quota,
			"calls": result.Calls,
			"yuan":  yuan,
		},
	})
}
