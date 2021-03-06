package xorm

import (
	"fmt"
	"sync"

	"github.com/stack-labs/stack-rpc/pkg/config"
	"github.com/stack-labs/stack-rpc/util/log"
	"xorm.io/xorm"
	"xorm.io/xorm/migrate"

	"github.com/stack-labs/starter-kit/console/account/conf"
)

var (
	dbConf conf.Database
	db     *xorm.Engine
	once   sync.Once
)

func InitDB() {
	once.Do(func() {
		dbConf = conf.Database{}
		// TODO stack-rpc
		conf, _ := config.NewConfig()
		err := conf.Get("database").Scan(&dbConf)
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("ConnMaxLifetime: %v", dbConf.ConnMaxLifetime)

		db, err = xorm.NewEngine("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
			dbConf.User,
			dbConf.Password,
			dbConf.Host,
			dbConf.Port,
			dbConf.Name,
		))
		if err != nil {
			log.Fatal(err)
		}

		db.SetMaxOpenConns(dbConf.MaxOpenConns)
		db.SetMaxIdleConns(dbConf.MaxIdleConns)
		db.SetConnMaxLifetime(dbConf.ConnMaxLifetime)

		// TODO xorm migrate问题，mysql创建migrations表出错
		// Specified key was too long; max key length is 767 bytes
		options := migrate.DefaultOptions
		exists, err := db.IsTableExist(options.TableName)
		if err != nil {
			panic(err)
		}
		if !exists {
			sql := fmt.Sprintf("CREATE TABLE %s (%s VARCHAR(64) PRIMARY KEY)", options.TableName, options.IDColumnName)
			if _, err := db.Exec(sql); err != nil {
				panic(err)
			}
		}

		m := migrate.New(db, migrate.DefaultOptions, migrations)
		err = m.Migrate()
		if err != nil {
			panic(err)
		}
	})
}
