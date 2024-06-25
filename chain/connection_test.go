package chain

import (
	log "github.com/ChainSafe/log15"
	"gotest.tools/assert"
	"testing"
)

func TestQueryKeysFile(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	api := conn.api
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		t.Fatal("get meta error")
	}
	key, err := generateFileKey(meta, []byte(TestCID))
	hash, err := api.RPC.Chain.GetBlockHash(15090000)
	strs, err := api.RPC.State.GetKeysPaged(PREFIX, 100, key, &hash)

	cids := make([]string, 0, len(strs))
	for _, str := range strs {
		println(str)
		cids = append(cids, parseCid(str))
	}
	//start := time.Now().UnixMilli()
	files, err := conn.GetFilesInfoV2ListWithCids(cids, &hash)
	//end := time.Now().UnixMilli()
	//println(end - start)
	assert.Equal(t, len(files), 100)
	//fmt.Printf("%v", fileV2)

}
