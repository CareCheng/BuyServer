package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

var orderCounter uint64

// GenerateOrderNo 生成订单号
// 格式：ORD_时间戳_序号
func GenerateOrderNo() string {
	timestamp := time.Now().Format("20060102150405")
	counter := atomic.AddUint64(&orderCounter, 1)
	return fmt.Sprintf("ORD_%s%04d", timestamp, counter%10000)
}

// GenerateLocalOrderNo 生成本地订单号
// 格式：ORD_时间戳_序号_随机字符串
func GenerateLocalOrderNo() string {
	timestamp := time.Now().Format("20060102150405")
	counter := atomic.AddUint64(&orderCounter, 1)

	// 生成4字节随机字符串
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("ORD_%s%04d_%s", timestamp, counter%10000, randomStr)
}

// ToDays 将时长转换为天数
func ToDays(value int, unit string) int {
	switch unit {
	case "天":
		return value
	case "周":
		return value * 7
	case "月":
		return value * 30
	case "年":
		return value * 365
	default:
		return value
	}
}
