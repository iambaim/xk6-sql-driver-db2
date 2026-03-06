// Package db2 contains IBM DB2 driver registration for xk6-sql.
package db2

import (
	dbsql "database/sql"

	"github.com/grafana/xk6-sql/sql"

	// Blank import required for initialization of driver.
	_ "github.com/ibmdb/go_ibm_db"
)

const driverName = "db2"

func init() {
	db, err := dbsql.Open("go_ibm_db", "")
	if err == nil {
		dbsql.Register(driverName, &db2Driver{inner: db.Driver()})
		db.Close()
	}

	sql.RegisterModule(driverName)
}
