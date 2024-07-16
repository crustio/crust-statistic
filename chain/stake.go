package chain

import (
	log "github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"sort"
)

const TotalStakePrefix = "0x5f3e4907f716ac89b6347d15ececedcad9489331c06779251388c89753b39481"

const StakeLimitPrefix = "0x5f3e4907f716ac89b6347d15ececedca2de7cf22faa2a2996bf0d7148f5375da"

const TB = 1 << 40

const CRU = 1e12

func GetTotalStakes(conn *connection) ([]Stake, error) {
	return getStakes(conn, TotalStakePrefix)
}

func GetStakeByIndex(conn *connection) (uint32, float64, error) {
	index, err := GetCurrentIndex(conn)
	if err != nil {
		return 0, 0, err
	}

	bytes, _ := types.EncodeToBytes(index)
	key, err := conn.generateKey("Staking", "ErasTotalStakes", bytes)
	if err != nil {
		return 0, 0, err
	}
	data, err := conn.GetStorageRawLatest(key)
	if err != nil {
		return 0, 0, err
	}
	cru, err := decodeCru(*data)
	if err != nil {
		return 0, 0, err
	}
	return index, cru, nil
}

func GetRewardByIndex(conn *connection) (uint32, float64, error) {
	index, err := GetCurrentIndex(conn)
	if err != nil {
		return 0, 0, err
	}
	index--
	bytes, _ := types.EncodeToBytes(index)
	key, err := conn.generateKey("Staking", "ErasStakingPayout", bytes)
	data, err := conn.GetStorageRawLatest(key)
	if err != nil {
		return 0, 0, err
	}
	cru, err := decodeCru(*data)
	if err != nil {
		return 0, 0, err
	}

	key, err = conn.generateKey("Staking", "ErasAuthoringPayout", bytes, []byte("0"))
	if err != nil {
		return 0, 0, err
	}
	prefix := types.HexEncodeToString(key)[:90]
	payout, err := getAuthoringPayout(conn, prefix)
	if v, ok := payout[index]; ok {
		cru += v
	}
	return index, cru, nil
}

func decodeCru(bs []byte) (float64, error) {
	var val types.U128
	err := types.DecodeFromBytes(bs, &val)
	if err != nil {
		return 0, err
	}
	return float64(val.Int64()) / float64(1e12), nil
}

func GetCurrentIndex(conn *connection) (uint32, error) {
	key, err := conn.generateKey("Staking", "CurrentEra")
	if err != nil {
		return 0, err
	}
	data, err := conn.GetStorageRawLatest(key)
	if err != nil {
		return 0, err
	}
	var val uint32
	err = types.DecodeFromBytes(*data, &val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func GetTopStakeLimit(conn *connection) ([]StakeLimit, error) {
	startKey := StakeLimitPrefix
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return nil, err
	}
	stakeSlice := make([]StakeLimit, 0, 100)
	for {
		keys, err := conn.api.RPC.State.GetKeysPaged(StakeLimitPrefix, 800, startKey, &hash)
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			break
		}
		// remove the startIndexKey
		if startKey == keys[0] {
			if len(keys) == 1 {
				break
			} else {
				keys = keys[1:]
			}
		}
		query := make([]types.StorageKey, 0, len(keys))
		for _, key := range keys {
			query = append(query, types.MustHexDecodeString(key))
		}
		resp, e := conn.QueryStorageAt(query, &hash)
		if e != nil {
			return nil, e
		}
		for _, set := range resp {
			for _, change := range set.Changes {
				var val types.U128
				err = types.DecodeFromBytes(change.StorageData, &val)
				if err != nil {
					continue
				}
				if val.Int64() > 0 {
					stakeSlice = append(stakeSlice, StakeLimit{
						Value: float64(val.Int64()),
						Acc:   encodeAccount(parseStakeAcc(change.StorageKey)),
					})
				}
			}
		}
		startKey = keys[len(keys)-1]
		log.Debug("get stake limit keys ", "cnt", len(keys))
	}
	if len(stakeSlice) > 0 {
		sort.Slice(stakeSlice, func(i, j int) bool {
			return stakeSlice[i].Value > stakeSlice[j].Value
		})
	}
	return stakeSlice[:10], nil
}

func GetStakingPayout(conn *connection) ([]Stake, error) {
	prefix := getPrefix("Staking", "ErasStakingPayout")
	return getStakes(conn, prefix)
}

func GetAuthoringPayout(conn *connection) (map[uint32]float64, error) {
	prefix := getPrefix("Staking", "ErasAuthoringPayout")
	return getAuthoringPayout(conn, prefix)
}

func getAuthoringPayout(conn *connection, prefix string) (map[uint32]float64, error) {
	startKey := prefix
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return nil, err
	}
	resMap := make(map[uint32]float64)
	for {
		keys, err := conn.api.RPC.State.GetKeysPaged(prefix, 1000, startKey, &hash)
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			break
		}
		// remove the startIndexKey
		if startKey == keys[0] {
			if len(keys) == 1 {
				break
			} else {
				keys = keys[1:]
			}
		}
		query := make([]types.StorageKey, 0, len(keys))
		for _, key := range keys {
			query = append(query, types.MustHexDecodeString(key))
		}
		resp, e := conn.QueryStorageAt(query, &hash)
		if e != nil {
			return nil, e
		}
		for _, set := range resp {
			for _, change := range set.Changes {
				var val types.U128
				err = types.DecodeFromBytes(change.StorageData, &val)
				if err != nil {
					continue
				}
				if val.Int64() > 0 {
					index := parseIndex(change.StorageKey)
					if v, ok := resMap[index]; ok {
						resMap[index] = v + float64(val.Int64())/1e12
					} else {
						resMap[index] = float64(val.Int64()) / 1e12
					}
				}
			}
		}
		startKey = keys[len(keys)-1]
	}
	return resMap, nil
}

func getStakes(conn *connection, prefix string) ([]Stake, error) {
	startKey := prefix
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return nil, err
	}
	keys, err := conn.api.RPC.State.GetKeysPaged(prefix, 1000, startKey, &hash)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	query := make([]types.StorageKey, 0, len(keys))
	for _, key := range keys {
		query = append(query, types.MustHexDecodeString(key))
	}
	resp, e := conn.QueryStorageAt(query, &hash)
	if e != nil {
		return nil, e
	}
	ss := make([]Stake, 0, len(keys))
	for _, set := range resp {
		for _, change := range set.Changes {
			val, err := decodeCru(change.StorageData)
			if err != nil {
				continue
			}
			index := parseIndex(change.StorageKey)
			ss = append(ss, Stake{
				Index: index,
				Value: val,
			})
		}
	}
	return ss, nil
}
