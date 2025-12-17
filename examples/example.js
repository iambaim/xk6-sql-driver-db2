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
