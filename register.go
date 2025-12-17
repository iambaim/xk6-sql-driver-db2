// Package db2 contains IBM DB2 driver registration for xk6-sql.
package db2

import (
	"github.com/grafana/xk6-sql/sql"

	// Blank import required for initialization of driver.
	_ "github.com/ibmdb/go_ibm_db"
)

func init() {
	sql.RegisterModule("go_ibm_db")
}
