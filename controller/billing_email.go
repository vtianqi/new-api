package controller

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// SendMonthlyBillEmail 给所有活跃用户发上月账单邮件
func SendMonthlyBillEmail() {
	now := time.Now()
	// 上个月的起止时间
	firstOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startTs := firstOfLastMonth.Unix()
	endTs := firstOfThisMonth.Unix()
	monthStr := firstOfLastMonth.Format("2006年01月")

	// 查询上月有调用的用户
	var userStats []struct {
		Username string `gorm:"column:username"`
		Quota    int64  `gorm:"column:quota"`
		Calls    int64  `gorm:"column:calls"`
	}
	model.LOG_DB.Table("logs").
		Select("username, sum(quota) as quota, count(*) as calls").
		Where("created_at >= ? AND created_at < ? AND type = 2", startTs, endTs).
		Group("username").
		Having("sum(quota) > 0").
		Scan(&userStats)

	if len(userStats) == 0 {
		common.SysLog("账单邮件：上月无活跃用户，跳过")
		return
	}

	common.SysLog(fmt.Sprintf("账单邮件：开始发送 %d 个用户的 %s 账单", len(userStats), monthStr))

	for _, stat := range userStats {
		// 查询用户邮箱
		var userEmail string
		model.DB.Table("users").Select("email").Where("username = ?", stat.Username).Scan(&userEmail)
		if userEmail == "" {
			continue
		}

		yuan := float64(stat.Quota) / common.QuotaPerUnit

		// 查询该用户上月各模型使用详情
		var modelDetails []struct {
			Model string `gorm:"column:model_name"`
			Quota int64  `gorm:"column:quota"`
			Calls int64  `gorm:"column:calls"`
		}
		model.LOG_DB.Table("logs").
			Select("model_name, sum(quota) as quota, count(*) as calls").
			Where("username = ? AND created_at >= ? AND created_at < ? AND type = 2",
				stat.Username, startTs, endTs).
			Group("model_name").
			Order("quota desc").
			Limit(5).
			Scan(&modelDetails)

		// 构建模型明细 HTML
		detailRows := ""
		for _, d := range modelDetails {
			dYuan := float64(d.Quota) / common.QuotaPerUnit
			detailRows += fmt.Sprintf(
				"<tr><td style='padding:6px 12px'>%s</td><td style='padding:6px 12px'>%d</td><td style='padding:6px 12px'>%.4f 元</td></tr>",
				d.Model, d.Calls, dYuan,
			)
		}

		subject := fmt.Sprintf("您的 %s API 使用账单", monthStr)
		content := fmt.Sprintf(`
<div style="font-family:sans-serif;max-width:600px;margin:0 auto">
  <h2>%s API 使用账单</h2>
  <p>您好 %s，以下是您 %s 的 API 使用情况：</p>
  <table style="width:100%%;border-collapse:collapse;margin:16px 0">
    <tr style="background:#f5f5f5">
      <td style="padding:8px 12px"><strong>总调用次数</strong></td>
      <td style="padding:8px 12px">%d 次</td>
    </tr>
    <tr>
      <td style="padding:8px 12px"><strong>总消费金额</strong></td>
      <td style="padding:8px 12px"><strong>%.4f 元</strong></td>
    </tr>
  </table>
  <h3>模型使用明细（Top 5）</h3>
  <table style="width:100%%;border-collapse:collapse;border:1px solid #eee">
    <tr style="background:#f5f5f5">
      <th style="padding:6px 12px;text-align:left">模型</th>
      <th style="padding:6px 12px;text-align:left">调用次数</th>
      <th style="padding:6px 12px;text-align:left">消费</th>
    </tr>
    %s
  </table>
  <p style="color:#888;margin-top:24px;font-size:13px">
    如有疑问请联系管理员。感谢您的使用！
  </p>
</div>
`, monthStr, stat.Username, monthStr, stat.Calls, yuan, detailRows)

		go func(email, subj, body string) {
			if err := common.SendEmail(subj, email, body); err != nil {
				common.SysError("发送账单邮件失败 " + email + ": " + err.Error())
			}
		}(userEmail, subject, content)
	}

	common.SysLog(fmt.Sprintf("账单邮件：%s 账单发送任务已启动", monthStr))
}
