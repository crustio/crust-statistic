package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
)

type sworkerMetrics struct {
}

func NewSworkerMetrics(cfg config.MetricConfig) sworkerMetrics {
	//prefix := ""
	//if strings.ToUpper(cfg.Env) == TEST {
	//	prefix = "Test_"
	//}
	return sworkerMetrics{}
}

func (s *sworkerMetrics) getSworkerCollector() []prometheus.Collector {
	return []prometheus.Collector{}
}
