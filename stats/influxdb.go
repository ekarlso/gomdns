package stats

import (
	log "code.google.com/p/log4go"
	"github.com/ekarlso/gomdns/config"
	influx "github.com/rcrowley/go-metrics/influxdb"
)

func setUpInflux(cfg *config.Configuration) {
	if cfg.InfluxUser == "" {
		log.Debug("No InfluxDB user specified, not pushing stats.")
		return
	}
	log.Debug("Setting up InfluxDB stats push")
	go influx.Influxdb(NameServerStats, 10e9, &influx.Config{
		Host:     cfg.InfluxHost,
		Database: cfg.InfluxDb,
		Username: cfg.InfluxUser,
		Password: cfg.InfluxPassword,
	})
}
