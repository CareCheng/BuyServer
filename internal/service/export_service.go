package service

import (
	"bytes"
	"fmt"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"

	"github.com/xuri/excelize/v2"
)

// ExportService 数据导出服务
// 支持订单、用户、日志等数据导出为Excel格式
type ExportService struct {
	repo *repository.Repository
}

// NewExportService 创建导出服务
func NewExportService(repo *repository.Repository) *ExportService {
	return &ExportService{repo: repo}
}

// ExportOrders 导出订单数据
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
//   - status: 订单状态（可选）
// 返回：
//   - Excel文件字节数据
//   - 错误信息
func (s *ExportService) ExportOrders(startDate, endDate time.Time, status string) ([]byte, error) {
	// 查询订单数据
	var orders []model.Order
	query := s.repo.GetDB().Model(&model.Order{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := "订单数据"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"订单号", "用户名", "商品名称", "单价", "时长", "时长单位", "状态", "支付方式", "卡密", "创建时间", "支付时间"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, order := range orders {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), order.OrderNo)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), order.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), order.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), order.Price)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), order.Duration)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), order.DurationUnit)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), getOrderStatusText(order.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), order.PaymentMethod)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), order.KamiCode)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), order.CreatedAt.Format("2006-01-02 15:04:05"))
		if order.PaymentTime != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), order.PaymentTime.Format("2006-01-02 15:04:05"))
		}
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	// 写入缓冲区
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportUsers 导出用户数据
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
// 返回：
//   - Excel文件字节数据
//   - 错误信息
func (s *ExportService) ExportUsers(startDate, endDate time.Time) ([]byte, error) {
	var users []model.User
	if err := s.repo.GetDB().Model(&model.User{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "用户数据"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"ID", "用户名", "邮箱", "邮箱已验证", "手机号", "状态", "两步验证", "注册时间", "最后登录"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, user := range users {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), user.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), boolToText(user.EmailVerified))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), user.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), getUserStatusText(user.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), boolToText(user.Enable2FA))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), user.CreatedAt.Format("2006-01-02 15:04:05"))
		if user.LastLoginAt != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), user.LastLoginAt.Format("2006-01-02 15:04:05"))
		}
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportOperationLogs 导出操作日志
// 参数：
//   - startDate: 开始日期
//   - endDate: 结束日期
//   - operatorType: 操作者类型（admin/user/空）
// 返回：
//   - Excel文件字节数据
//   - 错误信息
// 注意：日志已改为文件存储，此函数从文件读取日志数据
func (s *ExportService) ExportOperationLogs(startDate, endDate time.Time, userType string) ([]byte, error) {
	// 创建日志服务读取文件日志
	logSvc := NewLogService()
	
	// 获取日期范围内的所有日志
	var allLogs []LogEntry
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		logs, _, err := logSvc.GetOperationLogs(dateStr, 1, 10000, userType, "", "")
		if err != nil {
			continue
		}
		allLogs = append(allLogs, logs...)
	}

	f := excelize.NewFile()
	sheetName := "操作日志"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"ID", "用户类型", "用户ID", "用户名", "操作类型", "分类", "目标", "目标ID", "详情", "IP地址", "操作时间"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, log := range allLogs {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), log.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), log.UserType)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), log.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), log.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), log.Action)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), log.Target)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), log.TargetID)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), log.Detail)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), log.IP)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), log.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportLoginHistory 导出登录历史
// 参数：
//   - userID: 用户ID（0表示所有用户）
//   - startDate: 开始日期
//   - endDate: 结束日期
// 返回：
//   - Excel文件字节数据
//   - 错误信息
func (s *ExportService) ExportLoginHistory(userID uint, startDate, endDate time.Time) ([]byte, error) {
	var histories []model.LoginHistory
	query := s.repo.GetDB().Model(&model.LoginHistory{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate)
	
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	
	if err := query.Order("created_at DESC").Find(&histories).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "登录历史"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"ID", "用户ID", "用户名", "IP地址", "归属地", "设备", "浏览器", "操作系统", "状态", "失败原因", "登录时间"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, h := range histories {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), h.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), h.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), h.Username)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), h.IP)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), h.Location)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), h.Device)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), h.Browser)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), h.OS)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), getLoginStatusText(h.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), h.FailReason)
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), h.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ==================== 辅助函数 ====================

func getOrderStatusText(status int) string {
	switch status {
	case 0:
		return "待支付"
	case 1:
		return "已支付"
	case 2:
		return "已完成"
	case 3:
		return "已取消"
	case 4:
		return "已退款"
	default:
		return "未知"
	}
}

func getUserStatusText(status int) string {
	switch status {
	case 1:
		return "正常"
	case 0:
		return "禁用"
	default:
		return "未知"
	}
}

func getLoginStatusText(status int) string {
	if status == 1 {
		return "成功"
	}
	return "失败"
}

func boolToText(b bool) string {
	if b {
		return "是"
	}
	return "否"
}


// ExportUserOrders 导出用户自己的订单数据
// 参数：
//   - userID: 用户ID
// 返回：
//   - Excel文件字节数据
//   - 错误信息
func (s *ExportService) ExportUserOrders(userID uint) ([]byte, error) {
	var orders []model.Order
	if err := s.repo.GetDB().Model(&model.Order{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheetName := "我的订单"
	f.SetSheetName("Sheet1", sheetName)

	// 设置表头
	headers := []string{"订单号", "商品名称", "单价", "时长", "时长单位", "状态", "支付方式", "卡密", "创建时间", "支付时间"}
	for i, h := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, h)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%c1", 'A'+len(headers)-1), headerStyle)

	// 填充数据
	for i, order := range orders {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), order.OrderNo)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), order.ProductName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), order.Price)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), order.Duration)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), order.DurationUnit)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), getOrderStatusText(order.Status))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), order.PaymentMethod)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), order.KamiCode)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), order.CreatedAt.Format("2006-01-02 15:04:05"))
		if order.PaymentTime != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), order.PaymentTime.Format("2006-01-02 15:04:05"))
		}
	}

	// 自动调整列宽
	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheetName, col, col, 15)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
