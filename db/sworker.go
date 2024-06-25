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
