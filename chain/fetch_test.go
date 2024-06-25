package chain

import (
	"fmt"
	gsrpc "github.com/crustio/go-substrate-rpc-client/v4"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	scale "github.com/crustio/scale.go/types"
	"github.com/crustio/scale.go/types/scaleBytes"
	"github.com/crustio/scale.go/utiles"
	"gotest.tools/assert"
	"testing"
)

const TestKey = "0x5ebf094108ead4fefa73f7a3b13cb4a76ed21091d079415ef4a35264c626448d00047a4ec3aa6a75ec6261666b7265696479676e6b367535627663756336697661707378677770636b35716e627137377771747267747775377279793567346962616b75"
const TestCID = "bafkreidygnk6u5bvcuc6ivapsxgwpck5qnbq77wqtrgtwu7ryy5g4ibaku"
const TestUrl = "wss://rpc2-subscan.crust.network"

const TestCID2 = "QmPpivKGMs3BtszW9ekQKz3PsYJa9AvtqaoFBprJ3tzUbR"

var PREFIX = "0x5ebf094108ead4fefa73f7a3b13cb4a76ed21091d079415ef4a35264c626448d"

func TestParseCid(t *testing.T) {
	cid := parseCid(TestKey)
	if cid != TestCID {
		t.Fatal("cid error")
	}
}

func TestCidHex(t *testing.T) {
	bs := []byte(TestCID2)
	println(types.HexEncodeToString(bs))
}

func TestGenerateKey(t *testing.T) {
	api, _ := gsrpc.NewSubstrateAPI(TestUrl)
	hash, err := api.RPC.Chain.GetBlockHash(3979347)
	meta, err := api.RPC.State.GetMetadata(hash)
	if err != nil {
		t.Fatal("get meta error")
	}
	key, err := generateFileKey(meta, []byte(TestCID))
	if err != nil {
		t.Fatal("generate error", err)
	}
	assert.Equal(t, key, TestKey)
}

func TestQueryKey(t *testing.T) {
	api, _ := gsrpc.NewSubstrateAPI(TestUrl)
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		t.Fatal("get meta error")
	}
	key, err := generateFileKey(meta, []byte(TestCID))
	hash, err := api.RPC.Chain.GetBlockHash(15090000)
	strs, err := api.RPC.State.GetKeysPaged(PREFIX, 2, key, &hash)
	if err != nil {
		println(err.Error())
	}
	if len(strs) > 0 {
		for _, str := range strs {
			println(str)
		}
	}
	data, err := api.RPC.State.GetStorageRaw(types.MustHexDecodeString(key), hash)
	println(string(*data))
}

func TestQueryFile(t *testing.T) {
	api, err := gsrpc.NewSubstrateAPI(TestUrl)
	if err != nil {
		t.Fatal("connection error", err)
	}

	hash, err := api.RPC.Chain.GetBlockHash(15090000)

	//cid := parseCid(TestKey)
	data, err := api.RPC.State.GetStorageRaw(types.MustHexDecodeString(TestKey), hash)
	cid := parseCid(TestKey)
	println(cid)
	assert.Equal(t, cid, TestCID)
	if err != nil {
		println(err.Error())
	}

	m := scale.ScaleDecoder{}
	m.Init(scaleBytes.ScaleBytes{Data: *data}, nil)
	val := m.ProcessAndUpdateData("FileInfoV2<AccountId,Balance>")
	fileV2 := FileInfoV2{}
	utiles.UnmarshalAny(val, &fileV2)

	fmt.Printf(utiles.ToString(val))
	assert.Equal(t, len(fileV2.Replicas), 52)
	println("-------")
	//fmt.Printf("%v", fileV2)

}
