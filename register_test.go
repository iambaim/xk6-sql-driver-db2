package db2

import (
	_ "embed"
	"testing"

	"github.com/grafana/xk6-sql/sqltest"
)

//go:embed testdata/script.js
var script string

func TestDriverRegistration(t *testing.T) { //nolint:paralleltest
	sqltest.RunScript(t, driverName, "", `sql.open(driver, connection).close();`)
}

func TestIntegration(t *testing.T) { //nolint:paralleltest
	sqltest.RunScript(t, driverName, "HOSTNAME=localhost;DATABASE=sample;PORT=50000;UID=db2inst1;PWD=password123", script)
}
