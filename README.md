# xk6-sql-driver-db2

Database driver extension for [xk6-sql](https://github.com/grafana/xk6-sql) k6 extension to support IBM DB2 database.

## Example

```JavaScript file=examples/example.js
import sql from "k6/x/sql";
import driver from "k6/x/sql/driver/db2";

const con =
  "HOSTNAME=localhost;DATABASE=sample;PORT=50000;UID=db2inst1;PWD=password123";
const db = sql.open(driver, con);

export function setup() {
  let exist = db.query(
    "SELECT 1 FROM SYSCAT.TABLES WHERE TABSCHEMA='DB2INST1' AND TABNAME='SAMPLE';",
  );

  if (exist.length != 0) {
    db.exec("drop table SAMPLE;");
  }
  db.exec(`
     CREATE TABLE SAMPLE (
            id VARCHAR(10) NOT NULL DEFAULT '',
            f_name VARCHAR(40),
            l_name VARCHAR(40)
      );
  `);
}

export function teardown() {
  db.close();
}

export default function () {
  let result = db.exec(`
    INSERT INTO SAMPLE
      (id, f_name, l_name)
    VALUES
      ('1', 'Peter', 'Pan'),
      ('2', 'Wendy', 'Darling'),
      ('3', 'Tinker', 'Bell'),
      ('4', 'James', 'Hook');
  `);
  console.log(`${result.rowsAffected()} rows inserted`);

  let rows = db.query("SELECT * FROM SAMPLE WHERE f_name = 'Peter';");
  for (const row of rows) {
    console.log(row.L_NAME);
  }
}
```

## Building

This extension targets k6 v2. The build uses k6 v2.4.0 and xk6-sql v1.2.4. Scripts still import `k6/x/sql` and `k6/x/sql/driver/db2`.

Install Go 1.26 or later and xk6 v1.5.1. On Linux and macOS, you'll also need a C compiler for CGO. On Windows, run `make` from Git Bash; the driver loads the DB2 DLL at runtime without CGO.

On Linux, the IBM DB2 CLI driver needs `libxml2.so.2`. On Ubuntu 24.04, install it with `sudo apt-get install libxml2`.

Run `make build` to download the DB2 CLI driver and build k6. Run `make smoke-test` to check that k6 loads both SQL modules without connecting to DB2.

To run the integration test against a local DB2 container, run `make container-test` with Docker running and port 50000 free.

### Manual build

1. Download and extract DB2 CLI driver (available here: https://public.dhe.ibm.com/ibmdl/export/pub/software/data/db2/drivers/odbc_cli).
2. Install `xk6`:
    ```
    go install go.k6.io/xk6@v1.5.1
    ```
3. Set the environment variable, for example in Linux:
    ```
    export IBM_DB_HOME=<IBM DB2 CLI driver extracted path>
    export CGO_CFLAGS=-I${IBM_DB_HOME}/include
    export CGO_LDFLAGS=-L${IBM_DB_HOME}/lib
    export LD_LIBRARY_PATH=${IBM_DB_HOME}/lib
    export XK6_RACE_DETECTOR=1
    export CGO_ENABLED=1
    ```
4. Build this extension:
    ```
    xk6 build --k6-version v2.4.0 \
      --with github.com/grafana/xk6-sql@v1.2.4 \
      --with github.com/oleiade/xk6-encoding@v0.0.0-20260702145940-258b2345beb3 \
      --with github.com/iambaim/xk6-sql-driver-db2=.
    ```

## Usage

Check the [xk6-sql documentation](https://github.com/grafana/xk6-sql) on how to use this database driver.

## Publishing a release

Push a stable version tag (`vX.Y.Z`) from `main` to trigger the [Release workflow](https://github.com/iambaim/xk6-sql-driver-db2/actions/workflows/release.yml). It tests the tagged commit on Windows, macOS, and Ubuntu before publishing a GitHub Release with generated notes and source archives. The release does not include prebuilt binaries or IBM's DB2 CLI driver.

To publish a tag that already exists, such as `v0.2.0`, run the **Release** workflow from `main` and enter the tag. There's no need to move or recreate it.
