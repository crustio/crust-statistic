package metrics

import (
	"errors"
	"fmt"
	log "github.com/ChainSafe/log15"
	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"net/http"
	"statistic/config"
	"time"
)

const TEST = "TEST"

type ChainMetrics struct {
	fileMetrics
	sworkerMetrics
	stakeMetrics
	startCh   <-chan int
	stop      chan int
	config    config.MetricConfig
	scheduler *gocron.Scheduler
	pusher    *push.Pusher
}

var chainMetric *ChainMetrics

func NewChainMetrics(config *config.Config, startCh <-chan int) *ChainMetrics {

	chainMetric = &ChainMetrics{
		fileMetrics:    NewFileMetrics(config.Metric),
		sworkerMetrics: NewSworkerMetrics(config.Metric),
		stakeMetrics:   NewStakeMetrics(config.Metric),
		startCh:        startCh,
		stop:           make(chan int),
		config:         config.Metric,
		scheduler:      registerSecheduler(config.Metric),
	}
	chainMetric.registerMetric()
	initSlot(uint64(config.Chain.StartBlock))
	return chainMetric
}

func (c *ChainMetrics) registerMetric() {
	pusher := push.New(c.config.GateWay, "statistic-metric")
	pusher.Grouping("service", "statistic")
	register := prometheus.NewRegistry()
	register.MustRegister(c.getFileCollector()...)
	register.MustRegister(c.getSworkerCollector()...)
	register.MustRegister(c.getStakeCollector()...)
	pusher.Gatherer(register)
	c.pusher = pusher
	c.scheduler.Every(c.config.PushInterval).Seconds().Do(func() {
		time.Sleep(5 * time.Second)
		err := c.pusher.Add()
		if err != nil {
			log.Error("push metric err", "err", err)
		} else {
			log.Info("push metric success")
		}
	})
	prometheus.MustRegister(c.getFileCollector()...)
	prometheus.MustRegister(c.getSworkerCollector()...)
	prometheus.MustRegister(c.getStakeCollector()...)
}

func registerSecheduler(cfg config.MetricConfig) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	initHandler(cfg)
	for _, handler := range Handlers {
		s.Every(handler.interval).Seconds().Do(handler.handler)
	}
	return s
}

func (cm *ChainMetrics) Start() {
	cm.serve()
	go func() {
		err := cm.waitSchedule()
		if err != nil {
			return
		}
		log.Info("start metrics scheduler")
		cm.scheduler.StartAsync()
	}()
}

func (cm *ChainMetrics) waitSchedule() error {
	select {
	case <-cm.stop:
		return errors.New("chain metrics terminated")
	case <-cm.startCh:
		return nil
	}
}

func (cm *ChainMetrics) serve() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(fmt.Sprintf(":%d", cm.config.Port), nil)
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("Health status server is shutting down", err)
		} else {
			log.Error("Error serving metrics", "err", err)
		}
	}()
}

func (cm *ChainMetrics) Stop() {
	close(cm.stop)
	cm.scheduler.Stop()
}
