package chain

import (
	"statistic/db"
)

const (
	New int = iota
	UpdateBase
	UpdateRep
	Delete
)

func saveNewFile(fileInfo *FileInfoV2, cid string, number uint64) error {
	file := fileInfo.ToFileDto(cid, uint32(number))
	return db.SaveFiles(file, true)
}

func updateFileBase(fileInfo *FileInfoV2, cid string) error {
	file := fileInfo.ToFileSingleDto(cid)
	return db.UpdateFile(file)
}

func updateReplicas(fileInfo *FileInfoV2, cid string) error {
	file := fileInfo.ToFileDto(cid, 0)
	return db.UpdateReplicas(file)
}

func deleteByCid(cid string) error {
	return db.DeleteByCid(cid)
}
