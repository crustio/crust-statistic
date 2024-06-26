package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
)

var versionMap = map[string]string{
	"0xe6f4e6ab58d6ba4ba2f684527354156c009e4969066427ce18735422180b38f4": "v1.0.0",
	"0x673dcb16fe746ba752cd915133dc9135d59d6b7b022df58de2a7af4303fcb6e0": "v1.0.0",
	"0xff2c145fd797e1aef56b47a91adf3d3294c433bb29b035b3020d04a76200da0a": "v1.1.0",
	"0xa61ea2065a26a3f9f1e45ad02d8b2965c377b85ba409f6de7185c485d36dc503": "v1.1.1",
	"0x9469a2f6ea955d87d1b7296bf81e078898d31f0647b840389c184b206e51fc2d": "v1.1.1",
	"0x72041ba321cb982168beab2b3994f8b0b83a54e6dafaa95b444a3c273b490fb1": "v1.1.2",
	"0x69f72f97fc90b6686e53b64cd0b5325c8c8c8d7eed4ecdaa3827b4ff791694c0": "v2.0.0",
	"0x": "unknown",
}

type sworkerMetrics struct {
	storageSize                *prometheus.GaugeVec
	sworkerCnt                 *prometheus.GaugeVec
	sworkerCntByRatio          *prometheus.GaugeVec
	groupCnt                   *prometheus.GaugeVec
	avgSworkerCntByGroup       *prometheus.GaugeVec
	groupCntBySworkerCnt       *prometheus.GaugeVec
	groupCntByActiveSworkerCnt *prometheus.GaugeVec
	sworkerByVersion           *prometheus.GaugeVec
}

func NewSworkerMetrics(cfg config.MetricConfig) sworkerMetrics {
	if len(cfg.Codes) > 0 {
		for i, code := range cfg.Codes {
			versionMap[code] = cfg.Versions[i]
		}
	}
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
		sworkerByVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "SworkerByVersion",
				Help: "sworker number by version",
			},
			[]string{"version"},
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
		s.sworkerByVersion,
	}
}
