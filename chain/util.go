package chain

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"github.com/crustio/go-substrate-rpc-client/v4/xxhash"
	"github.com/crustio/scale.go/utiles"
	"github.com/decred/base58"
	"golang.org/x/crypto/blake2b"
	"regexp"
	"statistic/config"
)

const SlotSize = 600

func convertAccount(hex string) string {
	bytes := utiles.HexToBytes(hex)
	return SS58Encode(bytes, config.NetworkID)
}

func encodeAccount(bytes []byte) string {
	return SS58Encode(bytes, config.NetworkID)
}

func SS58Decode(address string) (uint16, []byte, error) {
	// Adapted from https://github.com/paritytech/substrate/blob/e6def65920d30029e42d498cb07cec5dd433b927/primitives/core/src/crypto.rs#L264

	data := base58.Decode(address)
	if len(data) < 2 {
		return 0, nil, fmt.Errorf("expected at least 2 bytes in base58 decoded address")
	}

	prefixLen := int8(0)
	ident := uint16(0)
	if data[0] <= 63 {
		prefixLen = 1
		ident = uint16(data[0])
	} else if data[0] < 127 {
		lower := (data[0] << 2) | (data[1] >> 6)
		upper := data[1] & 0b00111111
		prefixLen = 2
		ident = uint16(lower) | (uint16(upper) << 8)
	} else {
		return 0, nil, fmt.Errorf("invalid address")
	}

	checkSumLength := 2
	hash := ss58hash(data[:len(data)-checkSumLength])
	checksum := hash[:checkSumLength]

	givenChecksum := data[len(data)-checkSumLength:]
	if !bytes.Equal(givenChecksum, checksum) {
		return 0, nil, fmt.Errorf("checksum mismatch: expected %v but got %v", checksum, givenChecksum)
	}

	return ident, data[prefixLen : len(data)-checkSumLength], nil
}

func SS58Encode(pubkey []byte, format uint16) string {
	// Adapted from https://github.com/paritytech/substrate/blob/e6def65920d30029e42d498cb07cec5dd433b927/primitives/core/src/crypto.rs#L319
	ident := format & 0b0011_1111_1111_1111
	var prefix []byte
	if ident <= 63 {
		prefix = []byte{uint8(ident)}
	} else if ident <= 16_383 {
		// upper six bits of the lower byte(!)
		first := uint8(ident&0b0000_0000_1111_1100) >> 2
		// lower two bits of the lower byte in the high pos,
		// lower bits of the upper byte in the low pos
		second := uint8(ident>>8) | uint8(ident&0b0000_0000_0000_0011)<<6
		prefix = []byte{first | 0b01000000, second}
	} else {
		panic("unreachable: masked out the upper two bits; qed")
	}
	body := append(prefix, pubkey...)
	hash := ss58hash(body)
	return base58.Encode(append(body, hash[:2]...))
}

func ss58hash(data []byte) [64]byte {
	// Adapted from https://github.com/paritytech/substrate/blob/e6def65920d30029e42d498cb07cec5dd433b927/primitives/core/src/crypto.rs#L369
	prefix := []byte("SS58PRE")
	return blake2b.Sum512(append(prefix, data...))
}

func decodeCidsFromBlock(block *types.SignedBlock, index int) ([]string, error) {
	if len(block.Block.Extrinsics) <= index {
		return nil, errors.New("extrinsic out index")
	}
	ext := block.Block.Extrinsics[index]
	val, err := decodeCall(ext.Method.Args)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(val.Files))
	for _, file := range val.Files {
		res = append(res, string(file.Cid))
	}
	return res, nil
}

func decodeCidsFromSworkReport(block *types.SignedBlock, index int) ([]string, error) {
	if len(block.Block.Extrinsics) <= index {
		return nil, errors.New("extrinsic out index")
	}
	ext := block.Block.Extrinsics[index]
	val, err := decodeReportWork(ext.Method.Args)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(val.Add)+len(val.Del))
	for _, file := range val.Add {
		res = append(res, string(file.Cid))
	}
	for _, file := range val.Del {
		res = append(res, string(file.Cid))
	}
	return res, nil
}

func decodeReportWork(args types.Args) (*reportWork, error) {
	val := &reportWork{}
	err := types.DecodeFromBytes(args, val)
	if err != nil {
		return nil, err
	}
	return val, nil

}

func decodeCidsFromUpdateSpower(block *types.SignedBlock, index int) ([]string, error) {
	if len(block.Block.Extrinsics) <= index {
		return nil, errors.New("extrinsic out index")
	}
	ext := block.Block.Extrinsics[index]
	val, err := decodeUpdateSpower(ext.Method.Args)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(val.Files))
	for _, f := range val.Files {
		res = append(res, string(f.Cid))
	}
	return res, nil
}

func decodeUpdateSpower(args types.Args) (*updateSpower, error) {
	val := &updateSpower{}
	err := types.DecodeFromBytes(args, val)
	if err != nil {
		return nil, err
	}
	return val, nil

}

func decodeCall(bytes []byte) (*updateCall, error) {
	val := &updateCall{}
	err := types.DecodeFromBytes(bytes, val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func parseCid(key string) string {
	start := len(FileV2Prefix) + 18
	tmp := key[start:]
	bytes, err := types.HexDecodeString(tmp)
	if err != nil {
		return ""
	}
	newStr := string(bytes)
	re := regexp.MustCompile("/[^\\x00-\\x7F]/g")
	cid := re.ReplaceAllString(newStr, "")
	return cid
}

func parseAnchor(key string) string {
	start := len(SworkReportsPrefix) + 20
	tmp := key[start:]
	return "0x" + tmp
}

func parseHexAccountId(key string) string {
	tmp := key[66+32:]
	return "0x" + tmp
}

func parseAccountId(acc []byte) []byte {
	return acc[32+16:]
}

func parseStakeAcc(acc []byte) []byte {
	return acc[32+8:]
}

func parseIndex(key []byte) uint32 {
	tmp := key[32+8 : 44]
	var res uint32
	err := types.DecodeFromBytes(tmp, &res)
	if err != nil {
		return 0
	}
	return res
}

func generateFileKey(meta *types.Metadata, cid []byte) (string, error) {
	bytes, _ := types.EncodeToBytes(cid)
	storageKey, err := types.CreateStorageKey(meta, "Market", "FilesV2", bytes)
	if err != nil {
		return "", err
	}
	return types.HexEncodeToString(storageKey), nil
}

func getCidStorageKey(meta *types.Metadata, cid string) (types.StorageKey, error) {
	bytes, _ := types.EncodeToBytes([]byte(cid))
	return types.CreateStorageKey(meta, "Market", "FilesV2", bytes)
}

func getPrefix(prefix, method string) string {
	bs := append(xxhash.New128([]byte(prefix)).Sum(nil), xxhash.New128([]byte(method)).Sum(nil)...)
	return types.HexEncodeToString(bs)
}
