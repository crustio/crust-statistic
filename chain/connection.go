package chain

import (
	"github.com/ChainSafe/log15"
	gsrpc "github.com/crustio/go-substrate-rpc-client/v4"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	scale "github.com/crustio/scale.go/types"
	"github.com/crustio/scale.go/types/scaleBytes"
	"github.com/crustio/scale.go/utiles"
	"sync"
)

type connection struct {
	api      *gsrpc.SubstrateAPI
	log      log15.Logger
	url      string         // API endpoint
	meta     types.Metadata // Latest chain metadata
	metaLock sync.RWMutex   // Lock metadata for updates, allows concurrent reads
	stop     <-chan int     // Signals system shutdown, should be observed in all selects and loops
}

func NewConnection(url string, log log15.Logger, stop <-chan int) *connection {

	return &connection{url: url, log: log, stop: stop}
}

func (c *connection) getMetadata() *types.Metadata {
	c.metaLock.RLock()
	meta := c.meta
	c.metaLock.RUnlock()
	return &meta
}

func (c *connection) updateMetadata(hash *types.Hash) error {
	c.metaLock.Lock()
	meta, err := c.api.RPC.State.GetMetadata(*hash)
	if err != nil {
		c.metaLock.Unlock()
		return err
	}
	c.meta = *meta
	c.metaLock.Unlock()
	return nil
}

func (c *connection) Connect() error {
	c.log.Info("Connecting to substrate chain...", "url", c.url)
	api, err := gsrpc.NewSubstrateAPI(c.url)
	if err != nil {
		return err
	}
	opts := types.SerDeOptions{NoPalletIndices: true}
	types.SetSerDeOptions(opts)

	c.api = api

	// Fetch metadata
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
	}
	c.meta = *meta
	//c.log.Info("Fetched substrate metadata", "meta", meta)
	return nil
}

func (c *connection) GetEvents(meta *types.Metadata, hash *types.Hash) (*Events, error) {
	c.log.Trace("Fetching block for events", "hash", hash.Hex())
	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return nil, err
	}

	var records types.EventRecordsRaw
	_, err = c.api.RPC.State.GetStorage(key, &records, *hash)
	if err != nil {
		return nil, err
	}

	e := &Events{}
	err = records.DecodeEventRecords(meta, e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (c *connection) GetFilesInfoV2(cid string, hash *types.Hash) (*FileInfoV2, error) {
	key, err := c.generateKey([]byte(cid))
	if err != nil {
		return nil, err
	}
	data, err := c.GetStorageRaw(key, hash)
	if err != nil {
		return nil, err
	}

	m := scale.ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: *data}, nil)
	val := m.ProcessAndUpdateData("FileInfoV2<AccountId,Balance>")

	fileV2 := &FileInfoV2{}
	err = utiles.UnmarshalAny(val, fileV2)
	if err != nil {
		return nil, err
	}
	return fileV2, nil
}

func (c *connection) GetFilesInfoV2ListWithCids(cids []string, hash *types.Hash) ([]*StorageFile, error) {
	if len(cids) == 0 {
		return nil, nil
	}
	query := make([]types.StorageKey, 0, len(cids))
	keys := make(map[string]string)
	for _, cid := range cids {
		key, err := c.generateStorageKey(cid)
		if err != nil {
			return nil, err
		}
		query = append(query, key)
		keys[types.HexEncodeToString(key)] = cid
	}
	var number int
	sepSize := 200
	res := make([]*StorageFile, 0, len(cids))
	for len(query) > 0 {
		if len(query) > sepSize {
			number = sepSize
		} else {
			number = len(query)
		}
		subQuery := query[0:number]
		subRes, err := c.queryStorages(subQuery, hash, keys)
		if err != nil {
			return nil, err
		}
		res = append(res, subRes...)
		if len(query) > sepSize {
			query = query[number:]
		} else {
			break
		}
	}
	return res, nil
}

func (c *connection) GetFilesInfoV2ListWithKeys(hexKeys []string, hash *types.Hash) ([]*StorageFile, error) {
	if len(hexKeys) == 0 {
		return nil, nil
	}
	query := make([]types.StorageKey, 0, len(hexKeys))
	keys := make(map[string]string)
	for _, key := range hexKeys {
		cid := parseCid(key)
		query = append(query, types.MustHexDecodeString(key))
		keys[key] = cid
	}
	return c.queryStorages(query, hash, keys)
}

func (c *connection) queryStorages(query []types.StorageKey, hash *types.Hash, keys map[string]string) ([]*StorageFile, error) {
	resp, err := c.QueryStorageAt(query, hash)
	if err != nil {
		return nil, err
	}
	res := make([]*StorageFile, 0, len(query))
	for _, set := range resp {
		for _, change := range set.Changes {
			hexKey := types.HexEncodeToString(change.StorageKey)
			cid := keys[hexKey]
			m := scale.ScaleDecoder{}
			m.Init(scaleBytes.ScaleBytes{Data: change.StorageData}, nil)
			val := m.ProcessAndUpdateData("FileInfoV2<AccountId,Balance>")

			fileV2 := &FileInfoV2{}
			err = utiles.UnmarshalAny(val, fileV2)
			if err != nil {
				return nil, err
			}
			res = append(res, &StorageFile{
				cid,
				hexKey,
				fileV2,
			})
		}
	}
	return res, nil
}

func (c *connection) GetKeyPaged(prefix string, size uint32, startKey string, blockHash *types.Hash) ([]string, error) {
	return c.api.RPC.State.GetKeysPaged(prefix, size, startKey, blockHash)
}

func (c *connection) GetStorageRaw(key string, blockHash *types.Hash) (*types.StorageDataRaw, error) {
	return c.api.RPC.State.GetStorageRaw(types.MustHexDecodeString(key), *blockHash)
}

func (c *connection) GetBlock(hash *types.Hash) (*types.SignedBlock, error) {
	return c.api.RPC.Chain.GetBlock(*hash)
}

func (c *connection) QueryStorageAt(keys []types.StorageKey, blockHash *types.Hash) ([]types.StorageChangeSet, error) {
	return c.api.RPC.State.QueryStorageAt(keys, blockHash)
}

func (c *connection) GetLatestHeight() uint64 {
	header, err := c.api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		return 0
	}
	return uint64(header.Number)
}

func (c *connection) generateKey(cid []byte) (string, error) {
	return generateKey(&c.meta, cid)
}

func (c *connection) generateStorageKey(cid string) (types.StorageKey, error) {
	return getCidStorageKey(&c.meta, cid)
}
