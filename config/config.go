package config

import (
	log "github.com/ChainSafe/log15"
	"github.com/go-ini/ini"
	"github.com/urfave/cli/v2"
)

const DefaultConfigPath = "./config.ini"
const NetworkID = 66

type Config struct {
	Chain  ChainConfig
	Db     DbConfig
	Metric MetricConfig
}

type ChainConfig struct {
	Url             string
	StartBlock      uint64
	Size            uint64
	Confirm         int
	UseMarketUpdate bool
	UpdateSize      uint64
	ZeroNumber      uint64
}

type MetricConfig struct {
	GateWay         string
	Port            int
	Interval        int
	PushInterval    int
	StakeInterval   int
	SworkerInterval int
	Env             string
	Codes           []string
	Versions        []string
}

type DbConfig struct {
	Type        string
	User        string
	Password    string
	IP          string
	Port        string
	Name        string
	NumberShard int
}

func GetConfig(ctx *cli.Context) (*Config, error) {
	var fig Config
	path := DefaultConfigPath
	if file := ctx.String(ConfigFileFlag.Name); file != "" {
		path = file
	}
	err := loadConfig(path, &fig)
	if err != nil {
		log.Warn("err loading json file", "err", err.Error())
		return &fig, err
	}
	return &fig, nil
}

func loadConfig(filePath string, config *Config) error {
	cfg, err := ini.Load(filePath)
	if err != nil {
		log.Error("fail to loadConfig ", "filePath", filePath, "err", err.Error())
	}

	chain := ChainConfig{}
	cfg.Section("chain").MapTo(&chain)
	if err != nil {
		log.Error("load section error", "section", "chain", "error", err)
	}
	if chain.Size == 0 {
		chain.Size = 500000
	}

	if chain.UpdateSize == 0 {
		chain.UpdateSize = 100
	}
	db := DbConfig{}
	cfg.Section("db").MapTo(&db)
	if err != nil {
		log.Error("load section error", "section", "db", "error", err)
	}

	metric := MetricConfig{}
	cfg.Section("metric").MapTo(&metric)
	if err != nil {
		log.Error("load section error", "section", "metric", "error", err)
	}

	if metric.Interval == 0 {
		metric.Interval = 3600
	}
	if metric.SworkerInterval == 0 {
		metric.SworkerInterval = metric.Interval
	}

	if metric.StakeInterval == 0 {
		metric.StakeInterval = 3600 * 6
	}

	config.Chain = chain
	config.Db = db
	config.Metric = metric
	if len(metric.Codes) != len(metric.Versions) {
		panic("metric codes versions length error")
	}

	return nil
}
