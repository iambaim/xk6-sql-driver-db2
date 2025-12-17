package db2

import (
	_ "embed"
	"testing"

	"github.com/grafana/xk6-sql/sqltest"
)

//go:embed testdata/script.js
var script string

func TestIntegration(t *testing.T) { //nolint:paralleltest
	sqltest.RunScript(t, "go_ibm_db", "HOSTNAME=localhost;DATABASE=sample;PORT=50000;UID=db2inst1;PWD=password123", script)
}
