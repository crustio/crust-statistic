package main

import (
	log "github.com/ChainSafe/log15"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"statistic/chain"
	"statistic/config"
	"statistic/db"
	"statistic/metrics"
	"strconv"
	"syscall"
)

var app = cli.NewApp()

var (
	Version = "0.0.1"
)

var cliFlag = []cli.Flag{
	config.VerbosityFlag,
	config.ConfigFileFlag,
}

func init() {
	app.Action = run
	app.Copyright = "Copyright 2024 Crust Authors"
	app.Name = "Statistic"
	app.Usage = "Statistic"
	app.Authors = []*cli.Author{{Name: "Statistic 2024"}}
	app.Version = Version
	app.EnableBashCompletion = true
	app.Flags = append(app.Flags, cliFlag...)

}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startLogger(ctx *cli.Context) error {
	logger := log.Root()
	handler := logger.GetHandler()
	var lvl log.Lvl

	if lvlToInt, err := strconv.Atoi(ctx.String(config.VerbosityFlag.Name)); err == nil {
		lvl = log.Lvl(lvlToInt)
	} else if lvl, err = log.LvlFromString(ctx.String(config.VerbosityFlag.Name)); err != nil {
		return err
	}

	log.Root().SetHandler(log.LvlFilterHandler(lvl, handler))

	return nil
}

func run(ctx *cli.Context) error {
	err := startLogger(ctx)
	if err != nil {
		return err
	}

	log.Info("Starting Statistic...")

	cfg, err := config.GetConfig(ctx)
	if err != nil {
		return err
	}

	log.Debug("Config on initialization...", "config", *cfg)

	db.InitMysql(cfg.Db)

	logger := log.Root().New()

	chain, err := chain.NewChain(cfg.Chain, logger)
	if err != nil {
		return err
	}

	m := metrics.NewChainMetrics(cfg, chain.FetchCompleteCh())
	m.Start()
	chain.Start()

	// Used to signal core shutdown due to fatal error
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)
	// Block here and wait for a signal
	select {
	case <-sigc:
		logger.Warn("Interrupt received, shutting down now.")
	}
	m.Stop()
	chain.Stop()

	return nil
}
