# xk6-sql-driver-db2

Database driver extension for [xk6-sql](https://github.com/grafana/xk6-sql) k6 extension to support IBM DB2 database.

## Example

```JavaScript file=examples/example.js
import sql from "k6/x/sql";
import driver from "k6/x/sql/driver/go_ibm_db";
// Required as the DB2 driver seems to return uint arrays for VARCHAR columns
import { TextDecoder } from "k6/x/encoding";

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
  const streamDecoder = new TextDecoder();
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
    console.log(streamDecoder.decode(new Uint8Array(row.L_NAME)));
  }
}
```

## Building

The included `Makefile` is the preferred way to build this extension. Just execute `make k6`.

### Manual building step

1. Download and extract DB2 CLI driver (available here: https://public.dhe.ibm.com/ibmdl/export/pub/software/data/db2/drivers/odbc_cli).
2. Install `xk6`:
    ```
    go install go.k6.io/xk6@latest
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
4. Build this extension (we need encoding for query result texts):
    ```
    xk6 build --with github.com/grafana/xk6-sql@latest --with github.com/oleiade/xk6-encoding@latest --with github.com/iambaim/xk6-sql-driver-db2=.
    ```

## Usage

Check the [xk6-sql documentation](https://github.com/grafana/xk6-sql) on how to use this database driver.
