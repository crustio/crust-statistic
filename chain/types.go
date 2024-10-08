package chain

import (
	"math"
	"math/big"
	"statistic/db"

	"github.com/crustio/go-substrate-rpc-client/v4/types"
)

const ExpireDuration = 6 * 30 * 24 * 60 * 60 / 6

type FileInfoV2 struct {
	FileSize             uint64
	Spower               uint64
	ExpiredAt            uint32
	CalculatedAt         uint32
	CreateAt             uint32
	Amount               string
	Prepaid              string
	ReportedReplicaCount uint32
	RemainingPaidCount   uint32
	Replicas             map[string]Replica
}

type Replica struct {
	Who        string
	ValidAt    uint32
	Anchor     string
	IsReported bool
	CreateAt   uint32
}

type StorageFile struct {
	Cid  string
	Key  string
	File *FileInfoV2
}

func (f *FileInfoV2) ToFileDto(cid string, number uint32) *db.FileInfo {
	fileDto := &db.FileInfo{
		Cid:                cid,
		FileSize:           f.FileSize,
		Spower:             f.Spower,
		ExpiredAt:          f.ExpiredAt,
		CalculatedAt:       f.CalculatedAt,
		CreateAt:           number,
		Amount:             f.Amount,
		Prepaid:            f.Prepaid,
		ReportedReplicaCnt: f.ReportedReplicaCount,
		RemainingPaidCnt:   f.RemainingPaidCount,
	}
	replicaCreatAt := uint32(math.MaxUint32)
	if len(f.Replicas) > 0 {
		replicas := make([]db.Replica, 0, len(f.Replicas))
		for group, replica := range f.Replicas {
			replicas = append(replicas, db.Replica{
				GroupOwner: convertAccount(group),
				Who:        convertAccount(replica.Who),
				ValidAt:    replica.ValidAt,
				Anchor:     replica.Anchor,
				IsReported: replica.IsReported,
				CreateAt:   replica.CreateAt,
			})
			if replicaCreatAt > replica.CreateAt && replica.CreateAt > 0 {
				replicaCreatAt = replica.CreateAt
			}
		}
		fileDto.Replicas = replicas
	}
	if number == 0 {
		if f.ExpiredAt > ExpireDuration {
			number = f.ExpiredAt - ExpireDuration
		}
	}
	if number > replicaCreatAt && replicaCreatAt > 0 {
		number = replicaCreatAt
	}
	fileDto.CreateAt = number
	if f.Spower == 0 {
		fileDto.Spower = fileDto.FileSize
	}
	return fileDto
}

func (f *FileInfoV2) ToFileSingleDto(cid string) *db.FileInfo {
	fileDto := &db.FileInfo{
		Cid:                cid,
		FileSize:           f.FileSize,
		Spower:             f.Spower,
		ExpiredAt:          f.ExpiredAt,
		CalculatedAt:       f.CalculatedAt,
		Amount:             f.Amount,
		Prepaid:            f.Prepaid,
		ReportedReplicaCnt: f.ReportedReplicaCount,
		RemainingPaidCnt:   f.RemainingPaidCount,
	}
	if f.Spower == 0 {
		fileDto.Spower = fileDto.FileSize
	}
	return fileDto
}

type Erc721Token struct {
	Id       types.U256
	Metadata types.Bytes
}

type RegistryId types.H160
type TokenId types.U256

type SworkerPubKey []byte

type AssetId struct {
	RegistryId RegistryId
	TokenId    TokenId
}

type updateCall struct {
	Files      []filesInfo
	LastNumber uint32
}

type filesInfo struct {
	Cid      types.Bytes
	FileSize uint64
	Replicas []replicaExt
}

type replicaExt struct {
	Reporter     types.AccountID
	Owner        types.AccountID
	Anchor       types.Bytes
	Slot         uint64
	ReportNumber uint32
	ValidAt      uint32
	IsAdd        types.Bool
}

type reportWork struct {
	CurPk     SworkerPubKey
	UpgradePk SworkerPubKey
	Slot      types.U64
	SlotHash  types.Bytes
	SrdSize   types.U64
	FileSize  types.U64
	Add       []CidExt
	Del       []CidExt
	SrdRoot   types.Bytes
	FileRoot  types.Bytes
	Sig       types.Bytes
}

type CidExt struct {
	Cid types.Bytes
	V1  types.U64
	V2  types.U64
}

type workReport struct {
	Slot     uint64
	Spower   uint64
	Free     uint64
	FileSize uint64
	SrdRoot  types.Bytes
	FileRoot types.Bytes
	Anchor   string
}

func (w *workReport) ToDto() *db.WorkReport {
	reportDto := &db.WorkReport{
		Slot:     w.Slot,
		Spower:   w.Spower,
		Free:     w.Free,
		FileSize: w.FileSize,
		SrdRoot:  types.HexEncodeToString(w.SrdRoot),
		FileRoot: types.HexEncodeToString(w.FileRoot),
		Anchor:   w.Anchor,
	}
	if w.Free == 0 {
		reportDto.Ratio = 100
	} else {
		reportDto.Ratio = float64(reportDto.FileSize) / float64(reportDto.Free) * 100
	}
	return reportDto
}

type group struct {
	Members   []types.AccountID
	AllowList []types.AccountID
	GId       string
}

type identity struct {
	Anchor             []byte
	PunishmentDeadline uint64
	Group              types.OptionBytes32
}

func (w *group) ToDto(active db.GroupInfo) *db.SworkerGroup {
	return &db.SworkerGroup{
		GId:       w.GId,
		AllMember: len(w.Members),
		Active:    active.Active,
		Free:      active.FreeSum,
		Spower:    active.SpowerSum,
		FileSize:  active.FileSizeSum,
	}
}

func getInt64(val *big.Int) int64 {
	if val == nil {
		return 0
	}
	return val.Int64()
}

type pubInfo struct {
	Code   []byte
	Anchor types.OptionBytes
}

func (p *pubInfo) ToDto() *db.PubKey {
	anchor := "0x"
	ok, val := p.Anchor.Unwrap()
	if ok {
		anchor = types.HexEncodeToString(val)
	}
	return &db.PubKey{
		Code:   types.HexEncodeToString(p.Code),
		Anchor: anchor,
	}
}

type Stake struct {
	Index uint32
	Value float64
}

type StakeLimit struct {
	Acc   string
	Value float64
}

type updateSpower struct {
	Spowers []spower
	Files   []file
}

type spower struct {
	Anchor SworkerPubKey
	Number types.I64
}

type file struct {
	Cid        types.Bytes
	SpowerSize types.U64
	Replicas   []replicaArg
}

type replicaArg struct {
	Owner       types.AccountID
	Who         types.AccountID
	Anchor      SworkerPubKey
	BlockNumber types.BlockNumber
}

type calculateSpower struct {
	Cids []types.Bytes
}
