package crsql

import (
	"database/sql"
	"fmt"
)

type Change struct {
	Table      sql.NullString
	PK         sql.RawBytes
	CID        sql.NullString
	Val        *any
	ColVersion sql.NullInt64
	DBVersion  sql.NullInt64
	SiteID     sql.RawBytes
	CL         sql.NullInt64
	Seq        sql.NullInt64
}

func GetChanges(conn *sql.DB, currentVersion int) ([]Change, error) {
	// raw rows to insert into other database
	rs := make([]Change, 0)

	// we should be able to extract the changelog
	result, err := conn.Query("SELECT * FROM crsql_changes WHERE db_version > ? ", currentVersion)
	if err != nil {
		return rs, err
	}
	for result.Next() {
		ch := Change{}
		err := result.Scan(&ch.Table, &ch.PK, &ch.CID, &ch.Val, &ch.ColVersion, &ch.DBVersion,
			&ch.SiteID, &ch.CL, &ch.Seq)
		if err != nil {
			return rs, err
		}
		rs = append(rs, ch)
	}
	err = result.Close()
	if err != nil {
		return rs, err
	}
	return rs, nil
}

func MergeChanges(conn *sql.DB, changes []Change) error {
	tx, err := conn.Begin()
	fail := func(tx *sql.Tx, err error) error {
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("MergeChanges: %w", err)

		}
		return nil
	}
	if err != nil {
		return fail(tx, err)
	}
	for _, r := range changes {
		sql := `insert into crsql_changes
		("table", "pk", "cid", "val", "col_version", "db_version", "site_id", "cl", "seq") 
		values (?,?,?,?,?,?,?,?,?);`
		_, err := tx.Exec(sql, r.Table, r.PK, r.CID, r.Val, r.ColVersion, r.DBVersion, r.SiteID, r.CL, r.Seq)
		if err != nil {
			return fail(tx, err)
		}
	}
	return fail(tx, tx.Commit())
}
