package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

var ticketStatusMap = map[int]string{1: "待处理", 2: "处理中", 3: "已解决", 4: "已关闭"}

// CreateTicket 用户创建工单
func CreateTicket(c *gin.Context) {
	var req struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Priority int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数错误"})
		return
	}
	userId := c.GetInt("id")
	username := c.GetString("username")
	now := time.Now().Unix()

	ticket := &model.Ticket{
		UserId:      userId,
		Username:    username,
		Title:       req.Title,
		Status:      1,
		Priority:    max(1, req.Priority),
		CreatedTime: now,
		UpdatedTime: now,
	}
	if err := model.DB.Create(ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建失败"})
		return
	}
	// 保存第一条消息
	msg := &model.TicketMessage{
		TicketId:    ticket.Id,
		UserId:      userId,
		Username:    username,
		IsAdmin:     false,
		Content:     req.Content,
		CreatedTime: now,
	}
	model.DB.Create(msg)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": ticket})
}

// GetUserTickets 用户查自己的工单列表
func GetUserTickets(c *gin.Context) {
	userId := c.GetInt("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size := 20
	tickets, total, err := model.GetUserTickets(userId, (page-1)*size, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tickets, "total": total})
}

// GetAdminTickets 管理员查所有工单
func GetAdminTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size := 20
	tickets, total, err := model.GetAllTickets((page-1)*size, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tickets, "total": total})
}

// GetTicketDetail 查工单详情+消息
func GetTicketDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ticket, err := model.GetTicketById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "工单不存在"})
		return
	}
	// 非管理员只能看自己的工单
	userId := c.GetInt("id")
	isAdmin := c.GetBool("is_admin")
	if !isAdmin && ticket.UserId != userId {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "无权限"})
		return
	}
	msgs, _ := model.GetTicketMessages(id)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"ticket": ticket, "messages": msgs}})
}

// ReplyTicket 回复工单
func ReplyTicket(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct{ Content string `json:"content" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "内容不能为空"})
		return
	}
	ticket, err := model.GetTicketById(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "工单不存在"})
		return
	}
	userId := c.GetInt("id")
	username := c.GetString("username")
	isAdmin := c.GetBool("is_admin")
	if !isAdmin && ticket.UserId != userId {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "无权限"})
		return
	}
	now := time.Now().Unix()
	msg := &model.TicketMessage{
		TicketId: id, UserId: userId, Username: username,
		IsAdmin: isAdmin, Content: req.Content, CreatedTime: now,
	}
	model.DB.Create(msg)
	model.DB.Model(ticket).Update("updated_time", now)
	if isAdmin && ticket.Status == 1 {
		model.DB.Model(ticket).Update("status", 2)
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateTicketStatus 管理员更新工单状态
func UpdateTicketStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct{ Status int `json:"status"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Status < 1 || req.Status > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "状态值无效"})
		return
	}
	if err := model.DB.Model(&model.Ticket{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": req.Status, "updated_time": time.Now().Unix()}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": ticketStatusMap[req.Status]})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SendBillingEmail 手动触发账单邮件（管理员用）
func TriggerMonthlyBill(c *gin.Context) {
	go SendMonthlyBillEmail()
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "账单邮件发送任务已启动"})
}

// GetSystemStatus 系统状态接口（公开）
func GetSystemStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":  "operational",
			"version": common.Version,
		},
	})
}
