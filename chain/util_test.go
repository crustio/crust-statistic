package chain

import (
	gsrpc "github.com/crustio/go-substrate-rpc-client/v4"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"gotest.tools/assert"
	"testing"
)

func TestDecodeUpdateReplicas(t *testing.T) {
	files := []filesInfo{
		{
			[]byte("abc"),
			1,
			[]replicaExt{
				{
					types.AccountID{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x4, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d},
					types.AccountID{0xd4, 0x35, 0x93, 0xc7, 0x15, 0xfd, 0xd3, 0x1c, 0x61, 0x14, 0x1a, 0xbd, 0x4, 0xa9, 0x9f, 0xd6, 0x82, 0x2c, 0x85, 0x58, 0x85, 0x4c, 0xcd, 0xe3, 0x9a, 0x56, 0x84, 0xe7, 0xa5, 0x6d, 0xa2, 0x7d},
					[]byte("def"),
					1,
					2,
					3,
					true,
				},
			},
		},
	}
	num := uint32(666)
	var args []byte
	e, err := types.EncodeToBytes(files)
	if err != nil {
		t.Fatal(err)
	}
	args = append(args, e...)

	e, err = types.EncodeToBytes(num)
	if err != nil {
		t.Fatal(err)
	}
	args = append(args, e...)

	val, err := decodeCall(args)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(val.Files[0].Cid), "abc")
}

func Example_DecodeReportWork() {

	api, _ := gsrpc.NewSubstrateAPI(TestUrl)

	opts := types.SerDeOptions{NoPalletIndices: true}
	types.SetSerDeOptions(opts)

	hash, _ := api.RPC.Chain.GetBlockHash(15221812)
	block, _ := api.RPC.Chain.GetBlock(hash)

	cids, err := decodeCidsFromSworkReport(block, 1)
	if err != nil {
		panic(err)
	}
	for _, cid := range cids {
		println(types.HexEncodeToString([]byte(cid)))
	}
}
