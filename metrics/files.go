package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
	"strings"
)

type fileMetrics struct {
	filesCnt                prometheus.Gauge
	avgReplicas             prometheus.Gauge
	avgReplicasBySize       *prometheus.GaugeVec
	avgReplicasByCreateTime *prometheus.GaugeVec
	filesCntByReplicas      *prometheus.GaugeVec
	sumFileSpower           *prometheus.GaugeVec
	fileRatio               prometheus.Gauge
	fileCntBySlot           *prometheus.GaugeVec
	fileCntBySize           *prometheus.GaugeVec
	fileCntByCreateTime     *prometheus.GaugeVec
	fileCntByExpireTime     *prometheus.GaugeVec
}

func NewFileMetrics(cfg config.MetricConfig) fileMetrics {
	prefix := ""
	if strings.ToUpper(cfg.Env) == TEST {
		prefix = "Test_"
	}
	return fileMetrics{
		filesCnt: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "FilesCnt",
			Help: "Number of file",
		}),
		avgReplicas: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "AvgReplicas",
			Help: "Average Number of file replicas",
		}),
		avgReplicasBySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "AvgReplicasBySize",
				Help: "average number of file replicas by file size",
			},
			[]string{"size"},
		),
		avgReplicasByCreateTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "AvgReplicasByCreateTime",
				Help: "average number of file replicas by create time",
			},
			[]string{"createTime"},
		),
		filesCntByReplicas: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FilesCntByReplicas",
				Help: "Number of file by replica size",
			},
			[]string{"replicas"},
		),
		sumFileSpower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "SumFileSpower",
				Help: "File size and Spower(PB)",
			},
			[]string{"type"},
		),
		fileRatio: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "FileRatio",
			Help: "Ratio of spower to file size ",
		}),
		fileCntBySlot: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntBySlot",
				Help: "Number of file by slot",
			},
			[]string{"slot"},
		),
		fileCntBySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntBySize",
				Help: "Number of file by size",
			},
			[]string{"size"},
		),
		fileCntByCreateTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "FileCntByCreateTime",
				Help: "Number of file by create time",
			},
			[]string{"create"},
		),
		fileCntByExpireTime: prometheus.NewGaugeVec(
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
		f.avgReplicasByCreateTime,
		f.avgReplicasBySize,
		f.filesCnt,
		f.avgReplicas,
		f.filesCntByReplicas,
		f.sumFileSpower,
		f.fileRatio,
		f.fileCntBySlot,
		f.fileCntBySize,
		f.fileCntByCreateTime,
		f.fileCntByExpireTime,
	}
}
