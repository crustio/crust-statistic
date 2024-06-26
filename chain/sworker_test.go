package chain

import (
	"fmt"
	log "github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"gotest.tools/assert"
	"statistic/config"
	"statistic/db"
	"testing"
)

const GroupId = "cTLESxs4yM1nVZ1RPbXXYxfVuHPWHYmS4T94NDnbw3VycjBF7"

const AccountId = "cTHyvvRWJvZtvRwcYYip9X9JpnSWTbmvN6LTbeaDPsdNQ1YXY"

const MemberId = "cTGkqTGTysCxiPqERDyj5BcMxeeQgZGkT7E7GAMwCCzLaatry"

const anchor = "0xb4389d9844e4c9c8846f439f15e331fd8dc8cec752e27f643c969757be1cfb38d3d5188eb4b88b8d33d5608100964ed23aa4cef1845cbd90ecfe2c9a15ae9efb"

const ActiveAnchor = "0xbc46c627c337a2a4ed18645a8c2de7a49f7eb597e9543344b71c6c6d0a2331c8bafd32d2fca0a2c5f8f06ab232e62624c8b3e911dccd4d122440be5454cc90aa"

func getConfig() config.DbConfig {
	return config.DbConfig{
		Type:        "mysql",
		User:        "root",
		Password:    "admin",
		IP:          "10.230.255.15",
		Port:        "3306",
		Name:        "statistic",
		NumberShard: 4,
	}
}

func TestSworkKey(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	api := conn.api
	ab := types.MustHexDecodeString(anchor)
	anchorByte, _ := types.EncodeToBytes(ab)
	key, err := conn.generateKey("Swork", "WorkReports", anchorByte)
	an := parseAnchor(types.HexEncodeToString(key))
	println(an)

	if err != nil {
		panic(err)
	}
	println(types.HexEncodeToString(key))
	hash, err := api.RPC.Chain.GetBlockHash(15270000)
	data, err := api.RPC.State.GetStorageRaw(key, hash)

	val := &workReport{}
	err = types.DecodeFromBytes(*data, val)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", val)
	assert.Equal(t, val.Slot, uint64(4888800))
}

func TestGroups(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	//api := conn.api
	_, ab, _ := SS58Decode(GroupId)
	println(types.HexEncodeToString(ab))
	key, err := conn.generateKey("Swork", "Groups", ab)
	if err != nil {
		panic(err)
	}
	println(types.HexEncodeToString(key))
	println(GroupId)
	println(encodeAccount(parseAccountId(key)))
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	data, err := conn.api.RPC.State.GetStorageRaw(key, hash)
	fmt.Printf("%v", data)
	val := &group{}
	err = types.DecodeFromBytes(*data, val)
	if err != nil {
		panic(err)
	}
	println(len(val.Members))
}

func TestIdentityID(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	_, ab, _ := SS58Decode(MemberId)
	println(types.HexEncodeToString(ab))
	key, _ := conn.generateKey("Swork", "Identities", ab)
	an := parseHexAccountId(types.HexEncodeToString(key))
	println(an)
}

func TestGetGroups(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	db.InitMysql(getConfig())
	err := GetGroupInfo(conn)
	if err != nil {
		panic(err)
	}
}

func TestIdentity(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	//api := conn.api
	_, ab, _ := SS58Decode(AccountId)
	println(types.HexEncodeToString(ab))
	key, err := conn.generateKey("Swork", "Identities", ab)
	if err != nil {
		panic(err)
	}
	println(types.HexEncodeToString(key))
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	data, err := conn.api.RPC.State.GetStorageRaw(key, hash)
	fmt.Printf("%v\n", data)
	val := &identity{}
	err = types.DecodeFromBytes(*data, val)
	if err != nil {
		panic(err)
	}
	println(types.HexEncodeToString(val.Anchor))
}

func TestVersion(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	ab := types.MustHexDecodeString(ActiveAnchor)
	anchorByte, _ := types.EncodeToBytes(ab)
	key, err := conn.generateKey("Swork", "PubKeys", anchorByte)
	println(types.HexEncodeToString(key))
	if err != nil {
		panic(err)
	}
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	data, err := conn.api.RPC.State.GetStorageRaw(key, hash)
	val := &pubInfo{}
	err = types.DecodeFromBytes(*data, val)
	if err != nil {
		panic(err)
	}
	println(types.HexEncodeToString(val.Code))
	_, bs := val.Anchor.Unwrap()
	assert.Equal(t, ActiveAnchor, types.HexEncodeToString(bs))
	//fmt.Printf("%v", val)
}

func TestGetVersion(t *testing.T) {
	stop := make(chan int)
	conn := NewConnection(TestUrl, log.Root(), stop)
	conn.Connect()
	db.InitMysql(getConfig())
	err := GetPubKeys(conn)
	if err != nil {
		panic(err)
	}

}
