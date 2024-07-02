package chain

import (
	"errors"
	"github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"statistic/config"
	"statistic/db"
	"sync"
	"time"
)

const FileV2Prefix = "0x5ebf094108ead4fefa73f7a3b13cb4a76ed21091d079415ef4a35264c626448d"

const RetryCnt = 5

type fetcher struct {
	initBlock  uint64
	startBlock uint64
	log        log15.Logger
	stop       <-chan int
	completeCh chan int
	segfs      []*segFetcher
}

func NewFetcher(connections [3]*connection, cfg config.ChainConfig, startBlock uint64, logger log15.Logger, stop <-chan int) *fetcher {
	initBlock := cfg.StartBlock
	hash := fetchInit(connections[0], cfg.StartBlock)
	segfs := make([]*segFetcher, 0, 100)
	start := cfg.ZeroNumber
	end := start + cfg.Size - 1
	if end > initBlock {
		end = initBlock
	}
	for end < initBlock {
		segfs = append(segfs, newSegFetcher(connections, start, end, logger, hash, stop, cfg.UpdateSize))
		start = end + 1
		end = start + cfg.Size - 1
	}
	if end > initBlock {
		end = initBlock + 1
	}
	segfs = append(segfs, newSegFetcher(connections, start, end, logger, hash, stop, cfg.UpdateSize))

	return &fetcher{
		initBlock,
		startBlock,
		logger,
		stop,
		make(chan int),
		segfs,
	}
}

func (f *fetcher) start() {
	if f.startBlock > 0 {
		f.complete()
		return
	}
	var wg sync.WaitGroup
	wg.Add(len(f.segfs))
	for _, segf := range f.segfs {
		go segf.start(&wg)
		time.Sleep(time.Second)
	}
	wg.Wait()
	f.complete()
}

func fetchInit(conn *connection, endBlock uint64) *types.Hash {
	for {
		hash, err := conn.api.RPC.Chain.GetBlockHash(endBlock)
		if err != nil {
			time.Sleep(BlockRetryInterval)
			continue
		} else {
			return &hash
		}
	}
}

func (f *fetcher) complete() {
	f.log.Info("fetch done")
	close(f.completeCh)
}

func (f *fetcher) getCompleteCh() chan int {
	return f.completeCh
}

type segFetcher struct {
	conn       [3]*connection
	index      uint64
	end        uint64
	log        log15.Logger
	stop       <-chan int
	initHash   *types.Hash
	meta       *types.Metadata
	hashCh     chan *fileMeta
	fmCh       chan *fileMeta
	updateSize uint64
}

type fileMeta struct {
	blockNumber uint64
	hash        types.Hash
	cids        []string
}

func newSegFetcher(connection [3]*connection, index uint64, end uint64, logger log15.Logger, hash *types.Hash, stop <-chan int, updateSize uint64) *segFetcher {
	val, err := db.GetOrInit(index, end)
	if err != nil {
		panic(err)
	}
	index = val
	logger.Info("seg fetcher ", "start", index, "end", end)
	return &segFetcher{
		connection,
		index,
		end,
		logger,
		stop,
		hash,
		nil,
		make(chan *fileMeta, 10),
		make(chan *fileMeta, 10),
		updateSize,
	}
}

func (s *segFetcher) start(wg *sync.WaitGroup) {
	if s.index >= s.end {
		wg.Done()
		return
	}
	go func() {
		err := s.fetchHash(s.conn[0])
		if err != nil {
			s.log.Error("Fetch hash failed", "err", err)
		}
		close(s.hashCh)
	}()

	go func() {
		err := s.fetchFileMeta(s.conn[1])
		if err != nil {
			s.log.Error("Fetch event failed", "err", err)
		}
		close(s.fmCh)
	}()

	go func() {
		err := s.saveFiles(s.conn[2])
		if err != nil {
			s.log.Error("Fetch files failed", "err", err)
		}
		wg.Done()
	}()
}

func (s *segFetcher) fetchHash(conn *connection) error {
	indexNumber := s.index
	for {
		h, err := conn.api.RPC.Chain.GetBlockHash(indexNumber)
		if err != nil {
			s.log.Error("failed to get init block hash", "err", err)
			time.Sleep(time.Second)
			continue
		}
		meta, err := conn.api.RPC.State.GetMetadata(h)
		if err != nil {
			s.log.Error("failed to get init meta", "err", err)
			time.Sleep(time.Second)
			continue
		}
		s.meta = meta
		break
	}
	s.log.Info("seg fetcher complete  init", "end", s.end)
Main:
	for {
		select {
		case <-s.stop:
			return errors.New("terminated")
		default:
			// Get hash for index block, sleep and retry if not ready
			//now := time.Now().UnixMilli()
			hash, err := conn.api.RPC.Chain.GetBlockHash(indexNumber)
			//after := time.Now().UnixMilli()
			//f.log.Info("hash", "escape", after-now)
			if err != nil {
				s.log.Error("Failed to query block", "number", indexNumber, "err", err)
				time.Sleep(BlockRetryInterval)
				continue
			}
			s.log.Debug("get block hash", "number", indexNumber)
			fm := &fileMeta{
				blockNumber: indexNumber,
				hash:        hash,
			}
			s.hashCh <- fm
			indexNumber++
			if indexNumber > s.end {
				break Main
			}
		}
	}
	return nil
}

func (s *segFetcher) fetchFileMeta(conn *connection) error {
main:
	for {
		select {
		case <-s.stop:
			return errors.New("segFetcher terminated")
		case fm, ok := <-s.hashCh:
			// Get hash for index block, sleep and retry if not ready
			if ok {
				for {
					s.log.Debug("process block", "number", fm.blockNumber)
					err := s.processEvents(fm, conn)
					if err != nil {
						s.log.Error("Failed to process events in block", "number", fm.blockNumber, "err", err)
						continue
					}
					if fm.blockNumber == s.end {
						break main
					}
					break
				}
			} else {
				break main
			}
		}
	}
	return nil
}

func (s *segFetcher) processEvents(fm *fileMeta, conn *connection) error {
	//now := time.Now().UnixMilli()
	evts, err := conn.GetEvents(s.meta, &fm.hash)
	//after := time.Now().UnixMilli()
	//f.log.Info("event", "escape", after-now)
	if err != nil {
		return err
	}
	if len(evts.Market_FileSuccess) > 0 {
		cids := make([]string, 0, len(evts.Market_FileSuccess))
		for _, evt := range evts.Market_FileSuccess {
			s.log.Debug("get file success event", "cid", string(evt.Cid))
			if len(string(evt.Cid)) <= 64 {
				cids = append(cids, string(evt.Cid))
			} else {
				s.log.Error("get error event", "cid", string(evt.Cid), "block", fm.blockNumber)
			}

		}
		fm.cids = cids
	}
	s.fmCh <- fm
	if len(evts.System_CodeUpdated) > 0 {
		s.log.Trace("Received CodeUpdated event")
		meta, err := conn.api.RPC.State.GetMetadata(fm.hash)
		if err != nil {
			s.log.Error("Unable to update Metadata", "err", err)
		}
		s.meta = meta
	}

	s.log.Trace("Finished processing events", "block", fm.hash.Hex())
	return nil
}

func (s *segFetcher) saveFiles(conn *connection) error {
	nextUpdate := uint64(0)
main:
	for {
		select {
		case <-s.stop:
			return errors.New("segFetcher terminated")
		case fm, ok := <-s.fmCh:
			// Get hash for index block, sleep and retry if not ready
			if ok {
				s.saveKeys(fm, conn)
				if fm.blockNumber >= nextUpdate || fm.blockNumber == s.end {
					db.UpdateIndexKey(fm.blockNumber, s.end)
					nextUpdate = fm.blockNumber + s.updateSize
				}
			} else {
				break main
			}
		}
	}
	return nil
}

func (s *segFetcher) saveKeys(fm *fileMeta, conn *connection) {
	if len(fm.cids) == 0 {
		return
	}
	for {
		files, err := conn.GetFilesInfoV2ListWithCids(fm.cids, s.initHash)
		if err != nil {
			s.log.Error("segFetcher get file error", "number", fm.blockNumber, "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		for _, file := range files {
			dbFile := file.File.ToFileDto(file.Cid, uint32(fm.blockNumber))
			retry := RetryCnt
			for {
				err = db.SaveFiles(dbFile, false)
				if err != nil {
					retry--
					s.log.Error("save fileInfoV2 error", "cid", file.Cid, "err", err)
					if retry <= 0 {
						err = db.SaveError(&db.ErrorFile{
							Cid: file.Cid,
							Key: file.Key,
						})
						if err == nil {
							break
						}
					}
					time.Sleep(time.Second)
				} else {
					s.log.Info("save fileInfoV2 success", "cid", file.Cid)
					break
				}
			}
		}
		return
	}
}
