package db

import (
	"gorm.io/gorm"
	"strconv"
)

type CheckPoint struct {
	ID    int    `gorm:"primaryKey" json:"id"`
	Value string `gorm:"column:value;type:text"`
}

const (
	IndexKey         = 1
	IndexBlockNumber = 2
)

func GetIndexKey() (string, error) {
	var cp CheckPoint
	if err := MysqlDb.First(&cp, IndexKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			MysqlDb.Create(&CheckPoint{ID: IndexKey, Value: ""})
			return "", nil
		}
		return "", err
	}
	return cp.Value, nil
}

func UpdateIndexKey(value string) error {
	return MysqlDb.Model(&CheckPoint{}).Where("ID = ?", IndexKey).
		Update("value", value).Error
}

func GetBlockNumber() (uint64, error) {
	var cp CheckPoint
	if err := MysqlDb.First(&cp, IndexBlockNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			MysqlDb.Create(&CheckPoint{ID: IndexBlockNumber, Value: "0"})
			return 0, nil
		} else {
			return 0, err
		}
	}
	n, err := strconv.ParseInt(cp.Value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(n), nil
}

func UpdateBlockNumber(blockNumber uint64) error {
	val := strconv.FormatInt(int64(blockNumber), 10)
	return MysqlDb.Model(&CheckPoint{}).Where("ID = ?", IndexBlockNumber).
		Update("value", val).Error
}
