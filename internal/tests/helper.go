package tests

import (
	"fmt"

	"gorm.io/gorm/schema"

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

func DropTable(t schema.Tabler) error {
	db, err := mysql.New(&configs.MySQLConfig{DSN: DSN})
	if err != nil {
		return err
	}

	return db.Exec(fmt.Sprintf("DROP TABLE %s", t.TableName())).Error
}
