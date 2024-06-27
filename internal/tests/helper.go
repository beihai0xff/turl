package tests

import (
	"github.com/beihai0xff/turl/configs"
	"github.com/beihai0xff/turl/pkg/db/mysql"
)

func CreateTable(t any) error {
	db, err := mysql.New(&configs.MySQLConfig{DSN: DSN})
	if err != nil {
		return err
	}

	return db.AutoMigrate(&t)
}

func DropTable(t any) error {
	db, err := mysql.New(&configs.MySQLConfig{DSN: DSN})
	if err != nil {
		return err
	}

	return db.Migrator().DropTable(&t)
}
