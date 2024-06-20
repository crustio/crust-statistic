package chain

import (
	"errors"
	"github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"statistic/db"
	"time"
)

const FileV2Prefix = "0x5ebf094108ead4fefa73f7a3b13cb4a76ed21091d079415ef4a35264c626448d"

type fetcher struct {
	conn       *connection
	initBlock  uint64
	startBlock uint64
	startKey   string
	size       uint32
	log        log15.Logger
	stop       <-chan int
	completeCh chan int
	initHash   *types.Hash
	//initMeta   *types.Metadata
	fmCh chan []*StorageFile
}

func NewFetcher(connection *connection, size uint32, initBlock uint64, startBlock uint64, logger log15.Logger, stop <-chan int) *fetcher {
	key, err := db.GetIndexKey()
	if err != nil || key == "" {
		key = FileV2Prefix
	}
	return &fetcher{
		connection,
		initBlock,
		startBlock,
		key,
		size,
		logger,
		stop,
		make(chan int),
		nil,
		make(chan []*StorageFile, 1),
	}
}

func (f *fetcher) start() {
	if f.startBlock > 0 {
		f.complete()
		return
	}
	go func() {
		err := f.fetchKeys()
		if err != nil {
			f.log.Error("Fetch files failed", "err", err)
		}
	}()

	go func() {
		err := f.handlerFiles()
		if err != nil {
			f.log.Error("Fetch files failed", "err", err)
		}
	}()
}

func (f *fetcher) fetchInit() {
	for {
		hash, err := f.conn.api.RPC.Chain.GetBlockHash(f.initBlock)
		if err != nil {
			time.Sleep(BlockRetryInterval)
			f.log.Error("failed to get init block hash", "err", err)
			continue
		} else {
			f.initHash = &hash
			break
		}
	}
}

func (f *fetcher) fetchKeys() error {
	f.fetchInit()
	startIndexKey := f.startKey
	for {
		select {
		case <-f.stop:
			return errors.New("terminated")
		default:
			keys, err := f.conn.GetKeyPaged(FileV2Prefix, f.size, startIndexKey, f.initHash)
			if err != nil {
				f.log.Error("get keys error ", "starkey", startIndexKey, "err", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if len(keys) <= 1 {
				break
			}
			// remove the startIndexKey
			if startIndexKey == keys[0] {
				keys = keys[1:]
			}
			list, err := f.conn.GetFilesInfoV2ListWithKeys(keys, f.initHash)
			if err != nil {
				f.log.Error("get keys error ", "starkey", startIndexKey, "err", err)
				time.Sleep(5 * time.Second)
				continue
			}
			startIndexKey = keys[len(keys)-1]
			f.fmCh <- list
		}
	}
	f.complete()
	return nil
}

func (f *fetcher) handlerFiles() error {
	for {
		select {
		case <-f.stop:
			return errors.New("terminated")
		case files := <-f.fmCh:
			// Get hash for index block, sleep and retry if not ready
			f.saveFiles(files)
		}
	}
	return nil
}

func (f *fetcher) complete() {
	close(f.completeCh)
}

func (f *fetcher) getCompleteCh() chan int {
	return f.completeCh
}

func (f *fetcher) saveFiles(files []*StorageFile) {
	for _, file := range files {
		dbFile := file.File.ToFileDto(file.Cid, 0)
		err := db.SaveFiles(dbFile)
		if err != nil {
			f.log.Error("save fileInfoV2 error", "cid", file.Cid, "err", err)
		} else {
			f.log.Info("save fileInfoV2 success", "cid", file.Cid)
		}
	}
	f.startKey = files[len(files)-1].Key
	db.UpdateIndexKey(f.startKey)
}
