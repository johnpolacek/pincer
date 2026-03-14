package handlers

import (
	"github.com/boyter/pincer/common"
	"time"
)

type Timing struct {
	TimeMillis int64  `json:"timeMillis"`
	Source     string `json:"source"`
}

type HealthCheckResponse struct {
	Environment            *common.Environment `json:"environment"`
	MemoryStats            string              `json:"memoryStats"`
	MemoryAllocatedMb      uint64              `json:"memoryAllocatedMb"`
	StartTimeMs            int64               `json:"startTimeMs"`
	UptimeSeconds          int64               `json:"uptimeSeconds"`
	TotalActivity          int64               `json:"totalActivity"`
	IpCount                map[string]int64    `json:"ipCount"`
	TotalActivityProcessed int64               `json:"totalActivityProcessed"`
}

type HealthCheckResult struct {
	Success  bool                `json:"success"`
	Messages []string            `json:"messages"`
	Time     time.Time           `json:"time"`
	Timing   []Timing            `json:"timing"`
	Response HealthCheckResponse `json:"response"`
}
