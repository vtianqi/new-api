package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// DailyRevenue 每日营收数据
type DailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"` // 单位：元
	Quota   int64   `json:"quota"`
	Users   int64   `json:"users"`
	Calls   int64   `json:"calls"`
}

// ModelUsage 模型使用量
type ModelUsage struct {
	Model   string  `json:"model"`
	Calls   int64   `json:"calls"`
	Quota   int64   `json:"quota"`
	Revenue float64 `json:"revenue"`
	Percent float64 `json:"percent"`
}

// TopUser 活跃用户
type TopUser struct {
	Username string  `json:"username"`
	Calls    int64   `json:"calls"`
	Quota    int64   `json:"quota"`
	Revenue  float64 `json:"revenue"`
}

// GetRevenueStats 管理员：营收统计（每日趋势+模型分布+活跃用户）
func GetRevenueStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}

	// quota 换算比例（与 new-api 系统保持一致，500000 quota = 1元）
	const quotaPerYuan = 500000.0

	now := time.Now()
	startTime := now.AddDate(0, 0, -days).Unix()

	// 1. 每日营收趋势
	var dailyRows []struct {
		Date  string `gorm:"column:date"`
		Quota int64  `gorm:"column:quota"`
		Users int64  `gorm:"column:users"`
		Calls int64  `gorm:"column:calls"`
	}
	model.LOG_DB.Table("logs").
		Select("date(datetime(created_at, 'unixepoch', 'localtime')) as date, sum(quota) as quota, count(distinct username) as users, count(*) as calls").
		Where("created_at >= ? AND type = 2", startTime).
		Group("date").
		Order("date asc").
		Scan(&dailyRows)

	daily := make([]DailyRevenue, 0, len(dailyRows))
	for _, r := range dailyRows {
		daily = append(daily, DailyRevenue{
			Date:    r.Date,
			Revenue: float64(r.Quota) / quotaPerYuan,
			Quota:   r.Quota,
			Users:   r.Users,
			Calls:   r.Calls,
		})
	}

	// 2. 模型使用量排行
	var modelRows []struct {
		Model string `gorm:"column:model_name"`
		Calls int64  `gorm:"column:calls"`
		Quota int64  `gorm:"column:quota"`
	}
	model.LOG_DB.Table("logs").
		Select("model_name, count(*) as calls, sum(quota) as quota").
		Where("created_at >= ? AND type = 2", startTime).
		Group("model_name").
		Order("quota desc").
		Limit(10).
		Scan(&modelRows)

	var totalQuota int64
	for _, r := range modelRows {
		totalQuota += r.Quota
	}
	models := make([]ModelUsage, 0, len(modelRows))
	for _, r := range modelRows {
		pct := 0.0
		if totalQuota > 0 {
			pct = float64(r.Quota) / float64(totalQuota) * 100
		}
		models = append(models, ModelUsage{
			Model:   r.Model,
			Calls:   r.Calls,
			Quota:   r.Quota,
			Revenue: float64(r.Quota) / quotaPerYuan,
			Percent: pct,
		})
	}

	// 3. 活跃用户排行
	var userRows []struct {
		Username string `gorm:"column:username"`
		Calls    int64  `gorm:"column:calls"`
		Quota    int64  `gorm:"column:quota"`
	}
	model.LOG_DB.Table("logs").
		Select("username, count(*) as calls, sum(quota) as quota").
		Where("created_at >= ? AND type = 2", startTime).
		Group("username").
		Order("quota desc").
		Limit(10).
		Scan(&userRows)

	topUsers := make([]TopUser, 0, len(userRows))
	for _, r := range userRows {
		topUsers = append(topUsers, TopUser{
			Username: r.Username,
			Calls:    r.Calls,
			Quota:    r.Quota,
			Revenue:  float64(r.Quota) / quotaPerYuan,
		})
	}

	// 4. 汇总
	var summary struct {
		TotalQuota int64 `gorm:"column:total_quota"`
		TotalCalls int64 `gorm:"column:total_calls"`
		TotalUsers int64 `gorm:"column:total_users"`
	}
	model.LOG_DB.Table("logs").
		Select("sum(quota) as total_quota, count(*) as total_calls, count(distinct username) as total_users").
		Where("created_at >= ? AND type = 2", startTime).
		Scan(&summary)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"days":  days,
			"summary": gin.H{
				"total_revenue": float64(summary.TotalQuota) / quotaPerYuan,
				"total_quota":   summary.TotalQuota,
				"total_calls":   summary.TotalCalls,
				"total_users":   summary.TotalUsers,
			},
			"daily":     daily,
			"models":    models,
			"top_users": topUsers,
		},
	})
}

// GetUserRevenueSelf 用户端：自己的用量统计
func GetUserRevenueSelf(c *gin.Context) {
	username := c.GetString("username")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}

	const quotaPerYuan = 500000.0
	startTime := time.Now().AddDate(0, 0, -days).Unix()

	// 每日用量
	var dailyRows []struct {
		Date  string `gorm:"column:date"`
		Quota int64  `gorm:"column:quota"`
		Calls int64  `gorm:"column:calls"`
	}
	model.LOG_DB.Table("logs").
		Select("date(datetime(created_at, 'unixepoch', 'localtime')) as date, sum(quota) as quota, count(*) as calls").
		Where("created_at >= ? AND type = 2 AND username = ?", startTime, username).
		Group("date").
		Order("date asc").
		Scan(&dailyRows)

	daily := make([]gin.H, 0, len(dailyRows))
	for _, r := range dailyRows {
		daily = append(daily, gin.H{
			"date":    r.Date,
			"cost":    float64(r.Quota) / quotaPerYuan,
			"quota":   r.Quota,
			"calls":   r.Calls,
		})
	}

	// 模型分布
	var modelRows []struct {
		Model string `gorm:"column:model_name"`
		Calls int64  `gorm:"column:calls"`
		Quota int64  `gorm:"column:quota"`
	}
	model.LOG_DB.Table("logs").
		Select("model_name, count(*) as calls, sum(quota) as quota").
		Where("created_at >= ? AND type = 2 AND username = ?", startTime, username).
		Group("model_name").
		Order("quota desc").
		Limit(10).
		Scan(&modelRows)

	// 汇总
	var summary struct {
		TotalQuota int64 `gorm:"column:total_quota"`
		TotalCalls int64 `gorm:"column:total_calls"`
	}
	model.LOG_DB.Table("logs").
		Select("sum(quota) as total_quota, count(*) as total_calls").
		Where("created_at >= ? AND type = 2 AND username = ?", startTime, username).
		Scan(&summary)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"days": days,
			"summary": gin.H{
				"total_cost":  float64(summary.TotalQuota) / quotaPerYuan,
				"total_quota": summary.TotalQuota,
				"total_calls": summary.TotalCalls,
			},
			"daily":  daily,
			"models": modelRows,
		},
	})
}
