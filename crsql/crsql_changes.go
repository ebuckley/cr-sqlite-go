package crsql

import (
	"database/sql"
	"fmt"
	api "github.com/ebuckley/crsqlite-go/gen/api/v1"
)

//type Change struct {
//	Table      sql.NullString
//	PK         sql.RawBytes
//	CID        sql.NullString
//	Val        *any
//	ColVersion sql.NullInt64
//	DBVersion  sql.NullInt64
//	SiteID     sql.RawBytes
//	CL         sql.NullInt64
//	Seq        sql.NullInt64
//}

func GetChanges(conn *sql.DB, currentVersion int) ([]*api.Change, error) {
	// raw rows to insert into other database
	rs := make([]*api.Change, 0)

	// we should be able to extract the changelog
	result, err := conn.Query("SELECT * FROM crsql_changes WHERE db_version > ? ", currentVersion)
	if err != nil {
		return rs, err
	}
	for result.Next() {
		ch := api.Change{}
		err := result.Scan(&ch.Table, &ch.Pk, &ch.Cid, &ch.Val, &ch.ColVersion, &ch.DbVersion,
			&ch.SiteId, &ch.Cl, &ch.Seq)
		if err != nil {
			return rs, err
		}
		rs = append(rs, &ch)
	}
	err = result.Close()
	if err != nil {
		return rs, err
	}
	return rs, nil
}

func MergeChanges(conn *sql.DB, changes []*api.Change) error {
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
		_, err := tx.Exec(sql, r.Table, r.Pk, r.Cid, r.Val, r.ColVersion, r.DbVersion, r.SiteId, r.Cl, r.Seq)
		if err != nil {
			return fail(tx, err)
		}
	}
	return fail(tx, tx.Commit())
}
