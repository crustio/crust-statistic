package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
)

type sworkerMetrics struct {
	storageSize                *prometheus.GaugeVec
	sworkerCnt                 *prometheus.GaugeVec
	sworkerCntByRatio          *prometheus.GaugeVec
	groupCnt                   *prometheus.GaugeVec
	avgSworkerCntByGroup       *prometheus.GaugeVec
	groupCntBySworkerCnt       *prometheus.GaugeVec
	groupCntByActiveSworkerCnt *prometheus.GaugeVec
}

func NewSworkerMetrics(cfg config.MetricConfig) sworkerMetrics {
	return sworkerMetrics{
		storageSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "StorageSize",
				Help: "storage size (PB)",
			},
			[]string{"type"},
		),
		sworkerCnt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "SworkerCnt",
				Help: "sworker number",
			},
			[]string{"type"},
		),
		sworkerCntByRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "SworkerCntByRatio",
				Help: "sworker number by reported file size to free size",
			},
			[]string{"ratio"},
		),
		groupCnt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "GroupCnt",
				Help: "Group number",
			},
			[]string{"type"},
		),
		avgSworkerCntByGroup: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "AvgSworkerCntByGroup",
				Help: "average sworker number by group",
			},
			[]string{"type"},
		),
		groupCntBySworkerCnt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "GroupCntBySworkerCnt",
				Help: "group number by all sworker count",
			},
			[]string{"size"},
		),
		groupCntByActiveSworkerCnt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "GroupCntByActiveSworkerCnt",
				Help: "group number by active sworker count",
			},
			[]string{"size"},
		),
	}
}

func (s *sworkerMetrics) getSworkerCollector() []prometheus.Collector {
	return []prometheus.Collector{
		s.storageSize,
		s.sworkerCnt,
		s.sworkerCntByRatio,
		s.groupCnt,
		s.avgSworkerCntByGroup,
		s.groupCntBySworkerCnt,
		s.groupCntByActiveSworkerCnt,
	}
}
