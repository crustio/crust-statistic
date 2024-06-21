package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
	"strings"
)

type fileMetrics struct {
	FilesCnt                prometheus.Gauge
	AvgReplicas             prometheus.Gauge
	AvgReplicasBySize       *prometheus.GaugeVec
	AvgReplicasByCreateTime *prometheus.GaugeVec
	FilesCntByReplicas      *prometheus.GaugeVec
	SumFileSpower           *prometheus.GaugeVec
	FileRatio               prometheus.Gauge
	FileCntBySlot           *prometheus.GaugeVec
	FileCntBySize           *prometheus.GaugeVec
	FileCntByCreateTime     *prometheus.GaugeVec
	FileCntByExpireTime     *prometheus.GaugeVec
}

func NewFileMetrics(cfg config.MetricConfig) fileMetrics {
	prefix := ""
	if strings.ToUpper(cfg.Env) == TEST {
		prefix = "Test_"
	}
	return fileMetrics{
		FilesCnt: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "FilesCnt",
			Help: "Number of file",
		}),
		AvgReplicas: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "AvgReplicas",
			Help: "Average Number of file replicas",
		}),
		AvgReplicasBySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "AvgReplicasBySize",
				Help: "average number of file replicas by file size",
			},
			[]string{"size"},
		),
		AvgReplicasByCreateTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "AvgReplicasByCreateTime",
				Help: "average number of file replicas by create time",
			},
			[]string{"createTime"},
		),
		FilesCntByReplicas: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FilesCntByReplicas",
				Help: "Number of file by replica size",
			},
			[]string{"replicas"},
		),
		SumFileSpower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "SumFileSpower",
				Help: "File size and Spower(PB)",
			},
			[]string{"type"},
		),
		FileRatio: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "FileRatio",
			Help: "Ratio of spower to file size ",
		}),
		FileCntBySlot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntBySlot",
				Help: "Number of file by slot",
			},
			[]string{"slot"},
		),
		FileCntBySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntBySize",
				Help: "Number of file by size",
			},
			[]string{"size"},
		),
		FileCntByCreateTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntByCreateTime",
				Help: "Number of file by create time",
			},
			[]string{"create"},
		),
		FileCntByExpireTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntByExpireTime",
				Help: "Number of file by expire time",
			},
			[]string{"expire"},
		),
	}
}

func (f *fileMetrics) getFileCollector() []prometheus.Collector {
	return []prometheus.Collector{
		f.AvgReplicasByCreateTime,
		f.AvgReplicasBySize,
		f.FilesCnt,
		f.AvgReplicas,
		f.FilesCntByReplicas,
		f.SumFileSpower,
		f.FileRatio,
		f.FileCntBySlot,
		f.FileCntBySize,
		f.FileCntByCreateTime,
		f.FileCntByExpireTime,
	}
}
