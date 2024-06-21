package db

import (
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type FileInfo struct {
	ID                 int    `gorm:"primarykey" json:"id"`
	Cid                string `gorm:"unique;type:VARCHAR(64)"`
	FileSize           uint64 `gorm:"index:idx_file_size"`
	Spower             uint64
	ExpiredAt          uint32
	CreateAt           uint32 `gorm:"index:idx_file_create"`
	CalculatedAt       uint32
	Amount             string `gorm:"type:VARCHAR(128)"`
	Prepaid            string `gorm:"type:VARCHAR(128)"`
	ReportedReplicaCnt uint32 `gorm:"index:idx_file_replicas_cnt"`
	RemainingPaidCnt   uint32
	Replicas           []Replica `gorm:"-"`
}

type Replica struct {
	ID         int `gorm:"primarykey"`
	FileId     int
	GroupOwner string `gorm:"type:VARCHAR(64)"`
	Who        string `gorm:"type:VARCHAR(64)"`
	ValidAt    uint32
	Anchor     string `gorm:"type:VARCHAR(130)"`
	IsReported bool
	CreateAt   uint32
}

type ErrorFile struct {
	ID  int    `gorm:"primarykey"`
	Cid string `gorm:"unique;type:VARCHAR(64)"`
	Key string `gorm:"type:VARCHAR(300)"`
}

func SaveError(errFile *ErrorFile) error {
	err := MysqlDb.Create(errFile).Error
	if err != nil {
		if merr, ok := err.(*mysql.MySQLError); ok {
			if merr.Number != 1062 {
				return err
			} else {
				return nil
			}
		}
	}
	return nil
}

func SaveFiles(info *FileInfo) error {
	e := MysqlDb.Transaction(func(tx *gorm.DB) error {
		err := MysqlDb.Create(info).Error
		if err != nil {
			if merr, ok := err.(*mysql.MySQLError); ok {
				if merr.Number != 1062 {
					return err
				} else {
					return nil
				}
			}
		}
		for i := range info.Replicas {
			info.Replicas[i].FileId = info.ID
		}
		err = MysqlDb.CreateInBatches(&info.Replicas, len(info.Replicas)).Error
		return err
	})
	return e
}

func UpdateFile(info *FileInfo) error {
	return MysqlDb.Model(&FileInfo{}).Where("cid = ?", info.Cid).Updates(info).Error
}

func UpdateReplicas(info *FileInfo) error {
	file, err := QueryFileByCid(info.Cid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = MysqlDb.Create(info).Error
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		info.ID = file.ID
	}

	for i := range info.Replicas {
		info.Replicas[i].FileId = info.ID
	}
	err = MysqlDb.Transaction(func(tx *gorm.DB) error {
		e := DeleteReplicas(info.ID)
		if e != nil {
			return e
		}
		if e = MysqlDb.CreateInBatches(&info.Replicas, len(info.Replicas)).Error; e != nil {
			return e
		}

		return MysqlDb.Model(&FileInfo{}).
			Where("cid = ?", info.Cid).
			Update("reported_replica_cnt", info.ReportedReplicaCnt).
			Error
	})
	return err
}

func QueryFileByCid(cid string) (*FileInfo, error) {
	file := &FileInfo{}
	err := MysqlDb.Where("cid = ?", cid).First(file).Error
	if err != nil {
		return nil, err
	}
	return file, nil
}

func DeleteReplicas(fileId int) error {
	return MysqlDb.Delete(&Replica{}, "file_id = ?", fileId).Error
}

func DeleteByCid(cid string) error {
	file, err := QueryFileByCid(cid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	err = DeleteReplicas(file.ID)
	if err != nil {
		return err
	}
	return MysqlDb.Delete(file).Error
}

//func SaveReplica(re *Replica) error {
//	return MysqlDb.Save(re).Error
//}

func FileCnt() int64 {
	var count int64
	MysqlDb.Table("file_info").Count(&count)
	return count
}

func AvgReplicas() float64 {
	var avg float64
	MysqlDb.Table("file_info").
		Select("avg(reported_replica_cnt)").Scan(&avg)
	return avg
}

func AvgReplicasBySize(low uint64, high uint64) float64 {
	var avg float64
	MysqlDb.Table("file_info").
		Select("avg(reported_replica_cnt)").
		Where("file_size >= ?", low).
		Where("file_size < ?", high).Scan(&avg)
	return avg
}

func AvgReplicasByCreateTime(low uint64, high uint64) float64 {
	var avg float64
	MysqlDb.Table("file_info").
		Select("avg(reported_replica_cnt)").
		Where("create_at > ?", low).
		Where("create_at <= ?", high).Scan(&avg)
	return avg
}

func FileCntByReplicaSize(low uint64, high uint64) int64 {
	var count int64
	tx := MysqlDb.Table("file_info")
	if low == 0 && high == 0 {
		tx.Where("reported_replica_cnt = ?", 0)
	} else {
		tx.Where("reported_replica_cnt >= ?", low).
			Where("reported_replica_cnt <= ?", high)
	}
	tx.Count(&count)
	return count
}

func AvgFileSize() float64 {
	var avg float64
	MysqlDb.Table("file_info").
		Select("avg(file_size)").Scan(&avg)
	return avg
}

func AvgSpower() float64 {
	var avg float64
	MysqlDb.Table("file_info").
		Select("avg(spower)").Scan(&avg)
	return avg
}

func FileCntBySlot(slot uint64) (int64, error) {
	var count int64
	preSlot := slot - 600
	err := MysqlDb.Table("file_info").
		Where("create_at >= ?", preSlot).
		Where("create_at < ?", slot).Count(&count).Error
	return count, err
}

func FileCntBySize(low uint64, high uint64) int64 {
	var count int64
	MysqlDb.Table("file_info").
		Where("file_size >= ?", low).
		Where("file_size < ?", high).Count(&count)
	return count
}

func FileCntByCreateTime(low uint64, high uint64) int64 {
	var count int64
	MysqlDb.Table("file_info").
		Where("create_at > ?", low).
		Where("create_at <= ?", high).Count(&count)
	return count
}

func FileCntByExpireTime(low uint64, high uint64) int64 {
	var count int64
	MysqlDb.Table("file_info").
		Where("expired_at > ?", low).
		Where("expired_at <= ?", high).Count(&count)
	return count
}
