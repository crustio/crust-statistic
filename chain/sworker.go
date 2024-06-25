package chain

import (
	log "github.com/ChainSafe/log15"
	"github.com/crustio/go-substrate-rpc-client/v4/types"
	"statistic/db"
)

const SworkReportsPrefix = "0x2e3b7ab5757e6bbf28d3df3b5e01d6b9b7e949778e4650a54fcc65ad1f1ba39f"

const GroupPrefix = "0x2e3b7ab5757e6bbf28d3df3b5e01d6b92f583424865f2346c0f5c066f24dd499"

func GetGroupInfo(conn *connection) error {
	startKey := GroupPrefix
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return err
	}
	for {
		keys, err := conn.api.RPC.State.GetKeysPaged(GroupPrefix, 500, startKey, &hash)
		if err != nil {
			return err
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
		startKey = keys[len(keys)-1]
		query := make([]types.StorageKey, 0, len(keys))
		for _, key := range keys {
			query = append(query, types.MustHexDecodeString(key))
		}
		resp, e := conn.QueryStorageAt(query, &hash)
		if e != nil {
			return e
		}
		data := make(map[string]string)
		gs := make([]*Group, 0, len(keys))
		subQuery := make([]types.StorageKey, 0, 800)
		for _, set := range resp {
			for _, change := range set.Changes {
				val := &Group{}
				err = types.DecodeFromBytes(change.StorageData, val)
				if err != nil {
					return err
				}
				gs = append(gs, val)
				gid := encodeAccount(parseAccountId(change.StorageKey))
				val.GId = gid
				if len(val.Members) > 0 {
					for _, member := range val.Members {
						key, err := conn.generateKey("Swork", "Identities", member[:])
						if err != nil {
							return err
						}
						subQuery = append(subQuery, key)
					}
					if len(subQuery) > 600 {
						queryMember(subQuery, conn, &hash, data)
						subQuery = make([]types.StorageKey, 0, 800)
					}
				}
			}
		}
		if len(subQuery) > 0 {
			queryMember(subQuery, conn, &hash, data)
		}
		saveGroups(gs, data)
	}
	return nil
}

func saveGroups(groups []*Group, data map[string]string) error {
	dbg := make([]*db.SworkerGroup, 0, len(groups))
	var err error
	for _, group := range groups {
		var active int64 = 0
		if len(group.Members) > 0 {
			anchors := make([]string, 0, len(group.Members))
			for _, member := range group.Members {
				if v, ok := data[types.HexEncodeToString(member[:])]; ok {
					anchors = append(anchors, v)
				}
			}
			if len(anchors) > 0 {
				active, err = db.MemberCnt(anchors)
				if err != nil {
					return err
				}
			}
		}
		dbg = append(dbg, group.ToDto(int(active)))
	}
	return db.SaveGroups(dbg)
}

func queryMember(subQuery []types.StorageKey, conn *connection, hash *types.Hash, data map[string]string) error {
	log.Debug("query member", "count", len(subQuery))
	resp, e := conn.QueryStorageAt(subQuery, hash)
	if e != nil {
		return e
	}
	for _, set := range resp {
		for _, change := range set.Changes {
			val := &Identity{}
			err := types.DecodeFromBytes(change.StorageData, val)
			if err != nil {
				return err
			}
			data[parseHexAccountId(types.HexEncodeToString(change.StorageKey))] = types.HexEncodeToString(val.Anchor)
		}
	}
	log.Debug("query member done")
	return nil
}

func GetAllSworkReports(conn *connection) (int, int, error) {
	startKey := SworkReportsPrefix
	hash, err := conn.api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return 0, 0, err
	}
	head, err := conn.api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		return 0, 0, err
	}
	lastSlot := uint64(head.Number) / SlotSize * SlotSize
	activeSlot := lastSlot - 5*SlotSize
	allCount := 0
	activeCount := 0
	for {
		keys, err := conn.api.RPC.State.GetKeysPaged(SworkReportsPrefix, 500, startKey, &hash)
		if err != nil {
			return 0, 0, err
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
		allCount += len(keys)
		query := make([]types.StorageKey, 0, len(keys))
		for _, key := range keys {
			query = append(query, types.MustHexDecodeString(key))
		}
		resp, e := conn.QueryStorageAt(query, &hash)
		if e != nil {
			return 0, 0, e
		}
		res := make([]*db.WorkReport, 0, len(query))
		for _, set := range resp {
			for _, change := range set.Changes {
				val := &WorkReport{}
				err = types.DecodeFromBytes(change.StorageData, val)
				if err != nil {
					return 0, 0, err
				}
				if val.Slot < activeSlot {
					continue
				}
				hexKey := types.HexEncodeToString(change.StorageKey)
				workAnchor := parseAnchor(hexKey)
				val.Anchor = workAnchor
				res = append(res, val.ToDto())
			}
		}
		activeCount += len(res)
		err = db.SaveWorkReports(res)
		if err != nil {
			return 0, 0, err
		}
		startKey = keys[len(keys)-1]

	}
	return allCount, activeCount, nil
}
