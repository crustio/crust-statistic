package metrics

import (
	log "github.com/ChainSafe/log15"
	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"statistic/config"
	"strings"
	"time"
)

type stakeMetrics struct {
	cfg                  config.MetricConfig
	totalStakes          *prometheus.GaugeVec
	topStakeLimit        *prometheus.GaugeVec
	topValidatorSpower   *prometheus.GaugeVec
	topValidatorFileSize *prometheus.GaugeVec
	topValidatorRatio    *prometheus.GaugeVec
	guarantors           *prometheus.GaugeVec
	validators           *prometheus.GaugeVec
	rewards              *prometheus.GaugeVec
}

func NewStakeMetrics(cfg config.MetricConfig) stakeMetrics {
	prefix := ""
	if strings.ToUpper(cfg.Env) == TEST {
		prefix = "Test_"
	}
	return stakeMetrics{
		cfg: cfg,
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
			[]string{"eraIndex", "account"},
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
		guarantors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "StakeGuarantorCnt",
				Help: "Number of guarantors",
			},
			[]string{"eraIndex"},
		),
		validators: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prefix + "StakeValidatorCnt",
				Help: "Number of validators",
			},
			[]string{"eraIndex"},
		),
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

func (s *stakeMetrics) register(scheduler *gocron.Scheduler) {
	pusher := push.New(s.cfg.GateWay, "statistic-metric")
	pusher.Grouping("service", "statistic-stake")
	register := prometheus.NewRegistry()
	register.MustRegister(s.getStakeCollector()...)
	pusher.Gatherer(register)
	scheduler.Every(s.cfg.StakeInterval).Seconds().Do(func() {
		time.Sleep(60 * time.Second)
		err := pusher.Add()
		if err != nil {
			log.Error("push metric stake err", "err", err)
		} else {
			log.Info("push metric stake success")
		}
	})
}
