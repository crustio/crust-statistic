package db

import (
	"statistic/config"
	"testing"
)

func getConfig() config.DbConfig {
	return config.DbConfig{
		Type:        "mysql",
		User:        "root",
		Password:    "admin",
		IP:          "10.230.255.15",
		Port:        "3306",
		Name:        "statistic",
		NumberShard: 4,
	}
}

func TestDB(t *testing.T) {
	InitMysql(getConfig())
	rep := Replica{
		GroupOwner: "1",
	}
	file := FileInfo{
		Cid:                "QmcUhemhtzTvNtJR58yq2UBJ7tmDAzeaLTiQjxpixnx3oX",
		FileSize:           18416,
		ReportedReplicaCnt: 1,
		Replicas:           []Replica{rep},
	}
	UpdateFile(&file)
	UpdateReplicas(&file)

}

func TestCreateErr(t *testing.T) {
	InitMysql(getConfig())
	errfile := &ErrorFile{
		Cid: "1",
		Key: "0x",
	}
	err := SaveError(errfile)
	if err != nil {
		panic(err)
	}
}
