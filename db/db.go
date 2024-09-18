package db

import (
	_ "api-login/models"
	"api-login/nacos"
	"database/sql"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB(syncDB bool) {
	dbDriver := nacos.DBDriver
	dbUser := nacos.DBUser
	dbPass := nacos.DBPassword
	dbHost := nacos.DBHost
	dbPort := nacos.DBPort
	dbName := nacos.DBName

	// Construct data source name (DSN)
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8"

	// Register MySQL database driver
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// Register default database
	orm.RegisterDataBase("default", dbDriver, dsn)

	// Open a database connection
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		logs.Error("[InitDB] Open DB fail")
		return
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		logs.Error("[InitDB] Ping DB fail")
		return
	}
	aliasName := "default"

	err = orm.SetDataBaseTZ(aliasName, time.Local)
	if err != nil {
		logs.Error(err)
	}

	if syncDB {
		err = orm.RunSyncdb(aliasName, false, true)
		if err != nil {
			logs.Error(err)
		}
	}

	logs.Info("[InitDB] Init DB Success")
}

func GetDB() *sql.DB {
	return db
}
