package request

import (
	"AutoCert/src/utils"
	"fmt"
	"time"
)

func ScheduleDailyCheck(config utils.Config) {
	for {
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		duration := next.Sub(now)

		fmt.Printf("下次检查将在 %s 进行\n", next.Format("2006-01-02 15:04:05"))

		time.Sleep(duration)

		fmt.Println("开始执行每日证书检查...")
		utils.CheckSSLCertificates(config)
	}
}
