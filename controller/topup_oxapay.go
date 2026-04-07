package controller

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
)

// RequestOxaPayTopUp OxaPay 创建充值订单
func RequestOxaPayTopUp(c *gin.Context) {
	if !setting.OxaPayEnabled || setting.OxaPayApiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "OxaPay 未启用"})
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount < float64(setting.OxaPayMinTopUp) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("最低充值金额为 %d 元", setting.OxaPayMinTopUp)})
		return
	}

	userId := c.GetInt("id")
	username := c.GetString("username")
	orderId := fmt.Sprintf("oxapay_%d_%d", userId, time.Now().UnixMilli())

	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	baseUrl := scheme + "://" + c.Request.Host

	payload := map[string]interface{}{
		"merchant":       setting.OxaPayApiKey,
		"amount":         req.Amount,
		"currency":       "USD",
		"lifeTime":       30,
		"feePaidByPayer": 0,
		"description":    fmt.Sprintf("充值 %.2f 元 - %s", req.Amount, username),
		"orderId":        orderId,
		"callbackUrl":    baseUrl + "/api/oxapay/callback",
		"returnUrl":      baseUrl + "/console/topup",
	}

	payloadBytes, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", "https://api.oxapay.com/merchants/request", strings.NewReader(string(payloadBytes)))
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		common.SysError("OxaPay 请求失败: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "支付服务暂时不可用"})
		return
	}
	defer resp.Body.Close()

	var oxaResp struct {
		Result  int    `json:"result"`
		Message string `json:"message"`
		TrackId string `json:"trackId"`
		PayLink string `json:"payLink"`
	}
	json.NewDecoder(resp.Body).Decode(&oxaResp)

	if oxaResp.Result != 100 || oxaResp.PayLink == "" {
		common.SysError("OxaPay 响应异常: " + oxaResp.Message)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "创建支付订单失败: " + oxaResp.Message})
		return
	}

	quotaAmount := int64(req.Amount * common.QuotaPerUnit)
	topUp := &model.TopUp{
		UserId:        userId,
		Amount:        quotaAmount,
		Money:         req.Amount,
		TradeNo:       orderId,
		PaymentMethod: "OxaPay",
		CreateTime:    time.Now().Unix(),
		Status:        "pending",
	}
	model.DB.Create(topUp)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"pay_link": oxaResp.PayLink,
		"track_id": oxaResp.TrackId,
		"order_id": orderId,
	})
}

// OxaPayCallback OxaPay 回调处理
func OxaPayCallback(c *gin.Context) {
	body, _ := c.GetRawData()

	if setting.OxaPayCallbackKey != "" {
		signature := c.GetHeader("HMAC")
		mac := hmac.New(sha512.New, []byte(setting.OxaPayCallbackKey))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid signature"})
			return
		}
	}

	var callback struct {
		Status  string `json:"status"`
		OrderId string `json:"orderId"`
		TrackId string `json:"trackId"`
	}
	if err := json.Unmarshal(body, &callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}

	if callback.Status != "Paid" {
		c.JSON(http.StatusOK, gin.H{"message": "ignored"})
		return
	}

	var topUp model.TopUp
	if err := model.DB.Where("trade_no = ? AND status = ?", callback.OrderId, "pending").First(&topUp).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "order not found or already processed"})
		return
	}

	model.DB.Model(&topUp).Update("status", "success")
	err := model.IncreaseUserQuota(topUp.UserId, int(topUp.Amount), true)
	if err != nil {
		common.SysError(fmt.Sprintf("OxaPay 增加额度失败: userId=%d err=%v", topUp.UserId, err))
	} else {
		common.SysLog(fmt.Sprintf("OxaPay 充值成功: userId=%d orderId=%s", topUp.UserId, callback.OrderId))
		model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("通过 OxaPay 充值成功，金额: %d quota", topUp.Amount))
	}

	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
