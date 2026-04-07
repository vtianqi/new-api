package model

import (
	"gorm.io/gorm"
)

// Ticket 工单
type Ticket struct {
	Id          int            `json:"id" gorm:"primaryKey"`
	UserId      int            `json:"user_id" gorm:"index"`
	Username    string         `json:"username"`
	Title       string         `json:"title"`
	Status      int            `json:"status" gorm:"default:1"` // 1=待处理 2=处理中 3=已解决 4=已关闭
	Priority    int            `json:"priority" gorm:"default:1"` // 1=普通 2=紧急
	CreatedTime int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime int64          `json:"updated_time" gorm:"bigint"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TicketMessage 工单消息
type TicketMessage struct {
	Id          int   `json:"id" gorm:"primaryKey"`
	TicketId    int   `json:"ticket_id" gorm:"index"`
	UserId      int   `json:"user_id"`
	Username    string `json:"username"`
	IsAdmin     bool  `json:"is_admin"`
	Content     string `json:"content" gorm:"type:text"`
	CreatedTime int64 `json:"created_time" gorm:"bigint"`
}

func GetAllTickets(startIdx, num int) ([]*Ticket, int64, error) {
	var tickets []*Ticket
	var total int64
	DB.Model(&Ticket{}).Count(&total)
	err := DB.Order("updated_time desc").Offset(startIdx).Limit(num).Find(&tickets).Error
	return tickets, total, err
}

func GetUserTickets(userId, startIdx, num int) ([]*Ticket, int64, error) {
	var tickets []*Ticket
	var total int64
	DB.Model(&Ticket{}).Where("user_id = ?", userId).Count(&total)
	err := DB.Where("user_id = ?", userId).Order("updated_time desc").Offset(startIdx).Limit(num).Find(&tickets).Error
	return tickets, total, err
}

func GetTicketById(id int) (*Ticket, error) {
	var ticket Ticket
	err := DB.First(&ticket, id).Error
	return &ticket, err
}

func GetTicketMessages(ticketId int) ([]*TicketMessage, error) {
	var messages []*TicketMessage
	err := DB.Where("ticket_id = ?", ticketId).Order("created_time asc").Find(&messages).Error
	return messages, err
}
