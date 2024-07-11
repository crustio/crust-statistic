package chain

import (
	"errors"
	"github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"statistic/db"
	"time"
)

const BlockRetryInterval = time.Second * 5

type listener struct {
	conn       *connection
	startBlock uint64
	confirm    uint64
	log        log15.Logger
	stop       <-chan int
	completeCh <-chan int
	useMarket  bool
}

func NewListener(connection *connection, startBlock uint64, confirm uint64, logger log15.Logger, stop <-chan int, completeCh <-chan int, useMarket bool) *listener {
	return &listener{
		connection,
		startBlock,
		confirm,
		logger,
		stop,
		completeCh,
		useMarket,
	}
}

func (l *listener) start() {
	go func() {
		err := l.waitFetchComplete()
		if err != nil {
			return
		}
		err = l.pollBlocks()
		if err != nil {
			l.log.Error("Pool block failed", "err", err)
		}
	}()

}

func (l *listener) waitFetchComplete() error {
	select {
	case <-l.stop:
		return errors.New("listener terminated")
	case <-l.completeCh:
		return nil
	}
}

var ErrBlockNotReady = errors.New("required result to be 32 bytes, but got 0")

func (l *listener) pollBlocks() error {
	var currentBlock = l.startBlock
	var finalizedHeader *types.Header
	for {
		select {
		case <-l.stop:
			return errors.New("terminated")
		default:
			// No more retries, goto next block
			if finalizedHeader == nil || uint64(finalizedHeader.Number)-currentBlock < l.confirm {
				// Get finalized block hash
				finalizedHash, err := l.conn.api.RPC.Chain.GetFinalizedHead()
				if err != nil {
					l.log.Error("Failed to fetch finalized hash", "err", err)
					time.Sleep(BlockRetryInterval)
					continue
				}

				// Get finalized block header
				finalizedHeader, err = l.conn.api.RPC.Chain.GetHeader(finalizedHash)
				if err != nil {
					l.log.Error("Failed to fetch finalized header", "err", err)
					time.Sleep(BlockRetryInterval)
					continue
				}
			}

			// Sleep if the block we want comes after the most recently finalized block
			if currentBlock > uint64(finalizedHeader.Number) {
				l.log.Trace("Block not yet finalized", "target", currentBlock, "latest", finalizedHeader.Number)
				time.Sleep(BlockRetryInterval)
				continue
			}

			// Sleep if the difference is less than BlockDelay; (latest - current) < BlockDelay
			if uint64(finalizedHeader.Number)-currentBlock < l.confirm {
				l.log.Debug("Block not ready, will retry", "target", currentBlock, "latest", finalizedHeader.Number, "delay", l.confirm)
				time.Sleep(BlockRetryInterval)
				continue
			}

			// Get hash for latest block, sleep and retry if not ready
			hash, err := l.conn.api.RPC.Chain.GetBlockHash(currentBlock)
			if err != nil && err.Error() == ErrBlockNotReady.Error() {
				time.Sleep(BlockRetryInterval)
				continue
			} else if err != nil {
				l.log.Error("Failed to query block", "block", currentBlock, "err", err)
				time.Sleep(BlockRetryInterval)
				continue
			}
			l.log.Info("process block", "number", currentBlock)
			err = l.processEvents(&hash, currentBlock)
			if err != nil {
				l.log.Error("Failed to process events in block", "block", currentBlock, "err", err)
				continue
			}
			// Write block index
			currentBlock++
			l.startBlock = currentBlock
		}
	}
}

// processEvents fetches a block and parses out the events, calling listener.handleEvents()
func (l *listener) processEvents(hash *types.Hash, number uint64) error {
	events, err := l.conn.GetEvents(l.conn.getMetadata(), hash)
	if err != nil {
		return err
	}
	err = l.handleEvents(events, hash, number)
	if err != nil {
		return err
	}
	l.log.Trace("Finished processing events", "block", hash.Hex())
	return nil
}

// handleEvents calls the associated handler for all registered event types
func (l *listener) handleEvents(evts *Events, hash *types.Hash, number uint64) error {
	result := make(map[string]int)
	var err error

	for _, evt := range evts.Market_RenewFileSuccess {
		if _, ok := result[string(evt.Cid)]; !ok {
			result[string(evt.Cid)] = Update
		}
	}
	var block *types.SignedBlock
	if l.useMarket {
		for _, evt := range evts.Market_UpdateReplicasSuccess {
			if block == nil {
				block, err = l.conn.GetBlock(hash)
				if err != nil {
					return err
				}
			}
			cids, err := decodeCidsFromBlock(block, int(evt.Phase.AsApplyExtrinsic))
			if err != nil {
				return err
			}
			for _, cid := range cids {
				result[cid] = UpdateRep
			}
		}
	} else {
		for _, evt := range evts.Swork_WorksReportSuccess {
			if block == nil {
				block, err = l.conn.GetBlock(hash)
				if err != nil {
					return err
				}
			}
			cids, err := decodeCidsFromSworkReport(block, int(evt.Phase.AsApplyExtrinsic))
			if err != nil {
				return err
			}
			for _, cid := range cids {
				result[cid] = UpdateRep
			}
		}
	}

	for _, evt := range evts.Market_FileSuccess {
		result[string(evt.Cid)] = New
	}

	for _, evt := range evts.Market_IllegalFileClosed {
		result[string(evt.Cid)] = Delete
	}

	for _, evt := range evts.Market_FileClosed {
		result[string(evt.Cid)] = Delete
	}

	//update files with Cids
	if len(result) > 0 {
		err = l.updateFiles(result, hash, number)
		if err != nil {
			return err
		}
	}

	if len(evts.System_CodeUpdated) > 0 {
		l.log.Trace("Received CodeUpdated event")
		err := l.conn.updateMetadata(hash)
		if err != nil {
			l.log.Error("Unable to update Metadata", "error", err)
		}
	}

	err = db.UpdateBlockNumber(number)
	if err != nil {
		return err
	}

	return nil
}

func (l *listener) updateFiles(ops map[string]int, hash *types.Hash, number uint64) error {
	cids := make([]string, 0, len(ops))
	for key, op := range ops {
		if op != Delete {
			cids = append(cids, key)
		}
	}

	cidMap := make(map[string]*StorageFile)
	if len(cids) > 0 {
		files, err := l.conn.GetFilesInfoV2ListWithCids(cids, hash)
		if err != nil {
			return err
		}
		for _, file := range files {
			cidMap[file.Cid] = file
		}
	}
	var err error
	for cid, t := range ops {
		l.log.Info("handler file", "type", t, "cid", cid)
		//println(cid)
		switch t {
		case New:
			file, ok := cidMap[cid]
			if ok {
				err = saveNewFile(file.File, file.Cid, number)
			}
		case Update:
			file, ok := cidMap[cid]
			if ok {
				err = updateFile(file.File, file.Cid)
			}
		case UpdateRep:
			file, ok := cidMap[cid]
			if ok {
				err = updateReplicas(file.File, file.Cid)
			}
		case Delete:
			err = deleteByCid(cid)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
