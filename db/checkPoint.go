package db

import (
	"gorm.io/gorm"
)

type CheckPoint struct {
	ID        int `gorm:"primaryKey" json:"id"`
	CheckType int
	Value     uint64
	End       uint64
}

const (
	IndexKey         = 1
	IndexBlockNumber = 2
)

func GetOrInit(start, end uint64) (uint64, error) {
	var cp CheckPoint
	if err := MysqlDb.Where("check_type = ? and end = ? ", IndexKey, end).First(&cp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			MysqlDb.Create(&CheckPoint{CheckType: IndexKey, Value: start, End: end})
			return start, nil
		}
		return 0, err
	}
	return cp.Value, nil
}

func UpdateIndexKey(value, end uint64) error {
	return MysqlDb.Model(&CheckPoint{}).Where("check_type = ? and end = ?", IndexKey, end).
		Update("value", value).Error
}

func GetBlockNumber() (uint64, error) {
	var cp CheckPoint
	if err := MysqlDb.Where("check_type = ?  ", IndexBlockNumber).First(&cp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			MysqlDb.Create(&CheckPoint{CheckType: IndexBlockNumber, Value: 0})
			return 0, nil
		} else {
			return 0, err
		}
	}
	return cp.Value, nil
}

func UpdateBlockNumber(blockNumber uint64) error {
	return MysqlDb.Model(&CheckPoint{}).Where("check_type = ?", IndexBlockNumber).
		Update("value", blockNumber).Error
}
