package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/boyter/pincer/common"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

// HealthCheck is just a default health check endpoint but exposes some useful information
func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	log.Info().Str(common.UniqueCode, "d2b36a7").Str("ip", GetIP(r)).Msg("HealthCheck")

	start := common.MakeTimestampMilli()

	app.Service.ServiceMutex.RLock()
	defer app.Service.ServiceMutex.RUnlock()

	str := app.Service.StartTimeUnix
	upt := time.Now().Unix() - app.Service.StartTimeUnix

	_ipMapMutex.Lock()
	t, _ := json.MarshalIndent(HealthCheckResult{
		Success:  true,
		Messages: []string{},
		Time:     time.Now().UTC(),
		Timing: []Timing{
			{
				Source:     "HealthCheck",
				TimeMillis: common.MakeTimestampMilli() - start,
			},
		},
		Response: HealthCheckResponse{
			Environment:            app.Service.Environment,
			MemoryStats:            common.MemoryUsage(),
			MemoryAllocatedMb:      common.MemoryAllocatedMb(),
			StartTimeMs:            str,
			UptimeSeconds:          upt,
			TotalActivity:          app.Service.TotalActivity,
			IpCount:                _ipMap,
			TotalActivityProcessed: app.Service.TotalActivityProcessed,
		},
	}, "", jsonIndent)
	_ipMapMutex.Unlock()

	w.Header().Set("Content-Type", jsonContentType)
	_, _ = fmt.Fprint(w, string(t))
}
