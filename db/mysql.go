package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/sharding"
	"log"
	"statistic/config"
)

var MysqlDb *gorm.DB

func InitMysql(config config.DbConfig) {
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	dbUri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Asia%%2fShanghai&timeout=30s",
		config.User,
		config.Password,
		config.IP,
		config.Port,
		config.Name)

	dialector := mysql.New(mysql.Config{
		DSN:                       dbUri, // data source name
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	})
	var err error
	MysqlDb, err = gorm.Open(dialector, gormConfig)
	if nil != err {
		log.Fatalf("models.Setup err: %v", err)
	}

	middleware := sharding.Register(sharding.Config{
		ShardingKey:         "file_id",
		NumberOfShards:      uint(config.NumberShard),
		PrimaryKeyGenerator: sharding.PKSnowflake,
	}, "replica")
	MysqlDb.Use(middleware)

	if err = Migrator(); err != nil {
		log.Fatalf("migrator err: %v", err)
	}
}

func Migrator() error {
	if err := MysqlDb.Migrator().AutoMigrate(
		&CheckPoint{},
		&FileInfo{},
		&Replica{},
	); err != nil {
		return err
	}
	return nil
}
