package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"statistic/config"
	"strings"
)

type stakeMetrics struct {
	totalStakes          *prometheus.GaugeVec
	topStakeLimit        *prometheus.GaugeVec
	topValidatorSpower   *prometheus.GaugeVec
	topValidatorFileSize *prometheus.GaugeVec
	topValidatorRatio    *prometheus.GaugeVec
	guarantors           prometheus.Gauge
	validators           prometheus.Gauge
	rewards              *prometheus.GaugeVec
}

func NewStakeMetrics(cfg config.MetricConfig) stakeMetrics {
	prefix := ""
	if strings.ToUpper(cfg.Env) == TEST {
		prefix = "Test_"
	}
	return stakeMetrics{
		totalStakes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "TotalStakes",
				Help: "total stakes",
			},
			[]string{"eraIndex"},
		),
		topStakeLimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "TopStakeLimit",
				Help: "Top10 Stake Limit",
			},
			[]string{"account"},
		),
		topValidatorSpower: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "TopValidatorSpower",
				Help: "Top 70 validator spower",
			},
			[]string{"account"},
		),
		topValidatorFileSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "TopValidatorFileSize",
				Help: "Top 70 validator file size",
			},
			[]string{"account"},
		),
		topValidatorRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "TopValidatorRatio",
				Help: "Top 70 validator spower to free",
			},
			[]string{"account"},
		),
		guarantors: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "StakeGuarantorCnt",
			Help: "Number of guarantors",
		}),
		validators: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "StakeValidatorCnt",
			Help: "Number of validators",
		}),
		rewards: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "StakeRewards",
				Help: "Rewards by eraIndex",
			},
			[]string{"eraIndex"},
		),
	}
}

func (s *stakeMetrics) getStakeCollector() []prometheus.Collector {
	return []prometheus.Collector{
		s.totalStakes,
		s.topStakeLimit,
		s.topValidatorFileSize,
		s.topValidatorSpower,
		s.topValidatorRatio,
		s.guarantors,
		s.validators,
		s.rewards,
	}
}
