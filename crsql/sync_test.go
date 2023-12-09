package crsql

import (
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

var schema = `
create table foo (a primary key not null, b);
create table baz (a primary key not null, b, c, d);
select crsql_as_crr('foo');
select crsql_as_crr('baz');
`

func TestSimpleInsertMerge(t *testing.T) {
	conn, err := New(":memory:", schema)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Exec("insert into baz values (1, 'bthing', 'cthing', 'dthing')")
	if err != nil {
		t.Fatal(err)
	}

	rs, err := GetChanges(conn, 0)
	if err != nil {
		t.Fatal(rs)
	}

	// we should be able to apply the changelog to another database!

	otherDb, err := New(":memory:", schema)
	if err != nil {
		t.Fatal(err)
	}

	// check length of table is empty
	res, err := otherDb.Query("SELECT count(1) FROM baz")
	if err != nil {
		t.Fatal(err)
	}
	res.Next()
	var cnt int
	err = res.Scan(&cnt)
	if err != nil {
		t.Fatal(err)
	}
	res.Close()

	if cnt != 0 {
		t.Fatal("Should only have zero items in baz")
	}
	// --------------
	// start merging
	// --------------
	// start a transaction to insert the changes on the other database
	tx, err := otherDb.Begin()
	if err != nil {
		t.Fatal(err)
	}
	// now synchronize changes
	for _, r := range rs {
		sql := `insert into crsql_changes
		("table", "pk", "cid", "val", "col_version", "db_version", "site_id", "cl", "seq") 
		values (?,?,?,?,?,?,?,?,?);`
		_, err := tx.Exec(sql, r.Table, r.PK, r.CID, r.Val, r.ColVersion, r.DBVersion, r.SiteID, r.CL, r.Seq)
		if err != nil {
			t.Fatal(err)
		}
	}
	// eh...
	t.Log("inserted changes:", len(rs))
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	res, err = otherDb.Query("SELECT count(1) FROM baz")
	if err != nil {
		t.Fatal(err)
	}
	res.Next()
	err = res.Scan(&cnt)
	if err != nil {
		t.Fatal(err)
	}
	res.Close()

	if cnt != 1 {
		t.Fatal("Should only have one items in baz, got", cnt)
	}

	// TODO hmm!!!
	// they should be the same version now
	ogVer, err := GetDBVersion(conn)
	if err != nil {
		t.Fatal(err)
	}
	otherVer, err := GetDBVersion(otherDb)
	if err != nil {
		t.Fatal(err)
	}
	if ogVer != otherVer {
		t.Fatal("Both versions should be the same but we got ogVer", ogVer, "other", otherVer)
	}

	r, err := otherDb.Query("SELECT * FROM crsql_tracked_peers")
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
}
