package tests

import (
	"fmt"

	"gorm.io/gorm/schema"

	"github.com/beiai0xff/turl/configs"
	"github.com/beiai0xff/turl/pkg/db/mysql"
)

func CreateTable(t any) error {
	db, err := mysql.New(&configs.MySQLConfig{DSN: DSN})
	if err != nil {
		return err
	}

	return db.AutoMigrate(&t)
}

func DropTDDLTable(t schema.Tabler) error {
	db, err := mysql.New(&configs.MySQLConfig{DSN: DSN})
	if err != nil {
		return err
	}

	return db.Exec(fmt.Sprintf("DROP TABLE %s", t.TableName())).Error
}
