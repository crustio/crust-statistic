package chain

import (
	"errors"
	"github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"statistic/db"
	"sync"
	"time"
)

const FileV2Prefix = "0x5ebf094108ead4fefa73f7a3b13cb4a76ed21091d079415ef4a35264c626448d"

const RetryCnt = 5

type fetcher struct {
	conn       *connection
	initBlock  uint64
	startBlock uint64
	log        log15.Logger
	stop       <-chan int
	completeCh chan int
	segfs      []*segFetcher
}

func NewFetcher(connection *connection, size uint64, initBlock uint64, startBlock uint64, logger log15.Logger, stop <-chan int, updateSize uint64) *fetcher {
	hash := fetchInit(connection, initBlock)
	segfs := make([]*segFetcher, 0, 100)
	start := uint64(0)
	end := start + size - 1
	if end > initBlock {
		end = initBlock
	}
	for end < initBlock {
		segfs = append(segfs, newSegFetcher(connection, start, end, logger, hash, stop, updateSize))
		start = end + 1
		end = start + size - 1
	}
	if end > initBlock {
		end = initBlock + 1
	}
	segfs = append(segfs, newSegFetcher(connection, start, end, logger, hash, stop, updateSize))

	return &fetcher{
		connection,
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
	conn       *connection
	index      uint64
	end        uint64
	log        log15.Logger
	stop       <-chan int
	initHash   *types.Hash
	meta       *types.Metadata
	fmCh       chan *fileMeta
	updateSize uint64
}

type fileMeta struct {
	blockNumber uint64
	cids        []string
}

func newSegFetcher(connection *connection, index uint64, end uint64, logger log15.Logger, hash *types.Hash, stop <-chan int, updateSize uint64) *segFetcher {
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
		updateSize,
	}
}

func (s *segFetcher) start(wg *sync.WaitGroup) {
	if s.index >= s.end {
		wg.Done()
		return
	}
	go func() {
		err := s.fetchFileMeta()
		if err != nil {
			s.log.Error("Fetch files failed", "err", err)
		}
		close(s.fmCh)
	}()

	go func() {
		err := s.saveFiles()
		if err != nil {
			s.log.Error("Fetch files failed", "err", err)
		}
		wg.Done()
	}()
}

func (s *segFetcher) fetchFileMeta() error {
	indexNumber := s.index
	for {
		h, err := s.conn.api.RPC.Chain.GetBlockHash(indexNumber)
		if err != nil {
			s.log.Error("failed to get init block hash", "err", err)
			time.Sleep(time.Second)
			continue
		}
		meta, err := s.conn.api.RPC.State.GetMetadata(h)
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
			hash, err := s.conn.api.RPC.Chain.GetBlockHash(indexNumber)
			//after := time.Now().UnixMilli()
			//f.log.Info("hash", "escape", after-now)
			if err != nil {
				s.log.Error("Failed to query block", "number", indexNumber, "err", err)
				time.Sleep(BlockRetryInterval)
				continue
			}
			s.log.Info("process block", "number", indexNumber)
			err = s.processEvents(&hash, indexNumber)
			if err != nil {
				s.log.Error("Failed to process events in block", "number", indexNumber, "err", err)
				continue
			}
			indexNumber++
			if indexNumber > s.end {
				break Main
			}
		}
	}

	return nil
}

func (s *segFetcher) processEvents(hash *types.Hash, number uint64) error {
	//now := time.Now().UnixMilli()
	evts, err := s.conn.GetEvents(s.meta, hash)
	//after := time.Now().UnixMilli()
	//f.log.Info("event", "escape", after-now)
	if err != nil {
		return err
	}
	fm := &fileMeta{
		blockNumber: number,
	}
	if len(evts.Market_FileSuccess) > 0 {
		cids := make([]string, 0, len(evts.Market_FileSuccess))
		for _, evt := range evts.Market_FileSuccess {
			s.log.Info("get file success event", "cid", string(evt.Cid))
			cids = append(cids, string(evt.Cid))
		}
		fm.cids = cids
	}
	s.fmCh <- fm
	if len(evts.System_CodeUpdated) > 0 {
		s.log.Trace("Received CodeUpdated event")
		meta, err := s.conn.api.RPC.State.GetMetadata(*hash)
		if err != nil {
			s.log.Error("Unable to update Metadata", "err", err)
		}
		s.meta = meta
	}

	s.log.Trace("Finished processing events", "block", hash.Hex())
	return nil
}

func (s *segFetcher) saveFiles() error {
	nextUpdate := uint64(0)
main:
	for {
		select {
		case <-s.stop:
			return errors.New("segFetcher terminated")
		case fm, ok := <-s.fmCh:
			// Get hash for index block, sleep and retry if not ready
			if ok {
				s.saveKeys(fm)
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

func (s *segFetcher) saveKeys(fm *fileMeta) {
	if len(fm.cids) == 0 {
		return
	}
	for {
		files, err := s.conn.GetFilesInfoV2ListWithCids(fm.cids, s.initHash)
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
