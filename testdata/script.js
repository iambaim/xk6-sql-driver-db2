const db = sql.open(driver, connection);

let exist = db.query(
  "SELECT 1 FROM SYSCAT.TABLES WHERE TABSCHEMA='DB2INST1' AND TABNAME='TEST';",
);

if (exist.length != 0) {
  db.exec("drop table TEST;");
}

db.exec(
  "create table TEST(ID varchar(20),NAME varchar(20),LOCATION varchar(20),POSITION varchar(20));",
);

for (let i = 0; i < 5; i++) {
  db.exec(
    "INSERT INTO TEST (NAME, LOCATION, POSITION) VALUES ('name-" +
      i +
      "', 'location-" +
      i +
      "', 'position-" +
      i +
      "');",
  );
}

let all_rows = db.query("SELECT * FROM TEST;");
if (all_rows.length != 5) {
  throw new Error(
    "Expected all five rows to be returned; got " + all_rows.length,
  );
}

let one_row = db.query("SELECT * FROM TEST WHERE NAME = 'name-1';");
if (one_row.length != 1) {
  throw new Error("Expected single row to be returned; got " + one_row.length);
}

let no_rows = db.query("SELECT * FROM TEST WHERE NAME = 'bogus-name';");
if (no_rows.length != 0) {
  throw new Error("Expected no rows to be returned; got " + no_rows.length);
}

db.close();
