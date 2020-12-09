package metrics

import (
	metrics "github.com/rcrowley/go-metrics"
	"time"
)

var MetricsWriteBlockMeter = metrics.GetOrRegisterMeter("core.blockchain.writeBlock.time", nil)

// Config infos for influxdb
type Config struct {
	Addr     string        `json:"address"`
	Database string        `json:"database"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	Duration time.Duration `json:"duration"`
}
