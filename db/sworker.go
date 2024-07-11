package db

import (
	log "github.com/ChainSafe/log15"
	"gorm.io/gorm"
)

type WorkReport struct {
	ID       int    `gorm:"primarykey" json:"id"`
	Anchor   string `gorm:"index:idx_anchor;type:VARCHAR(130)"`
	Slot     uint64
	Spower   uint64
	Free     uint64
	FileSize uint64
	Ratio    float64 `gorm:"index:idx_ratio"`
	SrdRoot  string  `gorm:"type:VARCHAR(128)"`
	FileRoot string  `gorm:"type:VARCHAR(128)"`
}

func SaveWorkReports(reports []*WorkReport) error {
	e := MysqlDb.Transaction(func(tx *gorm.DB) error {
		err := MysqlDb.CreateInBatches(reports, 100).Error
		return err
	})
	return e
}

func ClearSworker() error {
	err := MysqlDb.Exec("truncate table sworker_group").Error
	if err != nil {
		return err
	}
	err = MysqlDb.Exec("truncate table pub_key").Error
	if err != nil {
		return err
	}
	return MysqlDb.Exec("truncate table work_report").Error
}

func SumFree() (float64, error) {
	var sum float64
	err := MysqlDb.Table("work_report").
		Select("sum(free)").Scan(&sum).Error
	return sum, err
}

func SumFileSize() (float64, error) {
	var sum float64
	err := MysqlDb.Table("work_report").
		Select("sum(file_size)").Scan(&sum).Error
	return sum, err
}

func NodeCntByRatio(low float64, high float64) (int64, error) {
	var count int64
	tx := MysqlDb.Table("work_report")
	if low == 0 && high == 0 {
		tx.Where("ratio = ?", 0)
	} else {
		tx.Where("ratio > ?", low).
			Where("ratio <= ?", high)
	}
	err := tx.Count(&count).Error
	return count, err
}

func MemberCnt(anchors []string) (int64, error) {
	var count int64
	err := MysqlDb.Table("work_report").Where("anchor in ?", anchors).Count(&count).Error
	return count, err
}

type SworkerGroup struct {
	ID        int    `gorm:"primarykey" json:"id"`
	GId       string `gorm:"type:VARCHAR(64)"`
	AllMember int    `gorm:"index:idx_all"`
	Active    int    `gorm:"index:idx_active"`
	Free      int64
	FileSize  int64
	Spower    int64
}

func SaveGroups(groups []*SworkerGroup) error {
	log.Debug("save groups", "cnt", len(groups))
	e := MysqlDb.Transaction(func(tx *gorm.DB) error {
		err := MysqlDb.CreateInBatches(groups, 100).Error
		return err
	})
	return e
}

func GroupCnt() (int64, error) {
	var count int64
	err := MysqlDb.Table("sworker_group").Count(&count).Error
	return count, err
}

func GroupActiveCnt() (int64, error) {
	var count int64
	err := MysqlDb.Table("sworker_group").Where("active > 0 ").Count(&count).Error
	return count, err
}

func AvgMembers() (float64, error) {
	var avg float64
	err := MysqlDb.Table("sworker_group").
		Select("avg(all_member)").Scan(&avg).Error
	return avg, err
}

func AvgActiveMembers() (float64, error) {
	var avg float64
	err := MysqlDb.Table("sworker_group").
		Select("avg(active)").Scan(&avg).Error
	return avg, err
}

func GroupCntByAll(low uint64, high uint64) (int64, error) {
	var count int64
	tx := MysqlDb.Table("sworker_group")
	if low == high {
		tx.Where("all_member = ?", low)
	} else {
		tx.Where("all_member >= ?", low).
			Where("all_member < ?", high)
	}
	err := tx.Count(&count).Error
	return count, err
}

func GroupCntByActive(low uint64, high uint64) (int64, error) {
	var count int64
	tx := MysqlDb.Table("sworker_group")
	if low == high {
		tx.Where("active = ?", low)
	} else {
		tx.Where("active >= ?", low).
			Where("active < ?", high)
	}
	err := tx.Count(&count).Error
	return count, err
}

func ActiveAnchors() ([]string, error) {
	var res []string
	err := MysqlDb.Raw("select anchor from work_report").Scan(&res).Error
	return res, err
}

type PubKey struct {
	ID     int    `gorm:"primarykey" json:"id"`
	Code   string `gorm:"type:VARCHAR(66)"`
	Anchor string `gorm:"index:idx_anchor;type:VARCHAR(130)"`
}

func SavePubKeys(keys []*PubKey) error {
	e := MysqlDb.Transaction(func(tx *gorm.DB) error {
		err := MysqlDb.CreateInBatches(keys, 100).Error
		return err
	})
	return e
}

type VersionCnt struct {
	Code string
	Cnt  int
}

func GetVersionCnt() ([]VersionCnt, error) {
	var res []VersionCnt
	err := MysqlDb.Raw("select pk.code,count(1) as cnt from work_report w left join pub_key pk on w.anchor = pk.anchor group by pk.code").
		Scan(&res).Error
	return res, err
}

type GroupInfo struct {
	Active      int
	SpowerSum   int64
	FreeSum     int64
	FileSizeSum int64
}

func GetGroupInfo(anchors []string) (GroupInfo, error) {
	var gi []GroupInfo
	err := MysqlDb.Raw("select count(1) as active,sum(spower) as spower_sum,sum(file_size) as file_size_sum,sum(free) as free_sum from work_report where anchor in ?",
		anchors).Scan(&gi).Error
	return gi[0], err
}

func GetTopGroups() ([]SworkerGroup, error) {
	var res []SworkerGroup
	err := MysqlDb.Table("sworker_group").Where("active > 0").
		Order("spower desc").Limit(70).Find(&res).Error
	return res, err
}
