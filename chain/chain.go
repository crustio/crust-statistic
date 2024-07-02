package chain

import (
	"github.com/ChainSafe/log15"
	"statistic/config"
	"statistic/db"
)

var DefaultConn *connection

type Chain struct {
	startBlock   uint64
	processBlock int
	startKey     string
	conn         *connection // THe chains connection
	fetcher      *fetcher
	listener     *listener
	stop         chan<- int
	logger       log15.Logger
}

func NewChain(cfg config.ChainConfig, logger log15.Logger) (*Chain, error) {

	stop := make(chan int)
	// Setup connection
	conns := [3]*connection{}
	for i := 0; i < 3; i++ {
		conn := NewConnection(cfg.Url, logger, stop)
		err := conn.Connect()
		if err != nil {
			return nil, err
		}
		conns[i] = conn
	}

	setDefaultConn(conns[0])
	initBlock := uint64(cfg.StartBlock)
	startBlock, err := db.GetBlockNumber()

	if err != nil {
		return nil, err
	}
	// Setup fetcher & listener
	f := NewFetcher(conns, cfg, startBlock, logger, stop)

	if startBlock == 0 {
		startBlock = initBlock + 1
	}
	l := NewListener(conns[0], startBlock, uint64(cfg.Confirm), logger, stop, f.getCompleteCh(), cfg.UseMarketUpdate)

	return &Chain{
		startBlock: cfg.StartBlock,
		conn:       conns[0],
		stop:       stop,
		logger:     logger,
		fetcher:    f,
		listener:   l,
	}, nil
}
func setDefaultConn(conn *connection) {
	DefaultConn = conn
}

func (c *Chain) Start() {
	c.fetcher.start()
	c.listener.start()
}

func (c *Chain) Stop() {
	close(c.stop)
}

func (c *Chain) FetchCompleteCh() <-chan int {
	return c.fetcher.getCompleteCh()
}
