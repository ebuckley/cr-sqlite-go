package crsql

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
)

func init() {
	sql.Register("sqlite3_crsql", &sqlite3.SQLiteDriver{
		Extensions: []string{
			"./crsqlite",
		},
	})
}

func New(file string, schema string) (*sql.DB, error) {
	connStr := fmt.Sprintf("file:%s?writable_schema=true", file)
	// we use plain old sql
	conn, err := sql.Open("sqlite3_crsql", connStr)
	if err != nil {
		return nil, err
	}
	if schema != "" {
		// we should be able to init a schema
		_, err = conn.Exec(schema)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}

func GetDBVersion(conn *sql.DB) (int, error) {
	var ver int
	r, err := conn.Query(`SELECT crsql_db_version()`)
	if err != nil {
		return -1, err
	}
	r.Next()
	err = r.Scan(&ver)
	if err != nil {
		return -1, err
	}
	err = r.Close()
	return ver, err
}

func GetPeerID(conn *sql.DB) ([]byte, error) {
	r, err := conn.Query("SELECT crsql_site_id")
	if err != nil {
		return nil, err
	}
	r.Next()
	var res []byte
	err = r.Scan(&res)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return res, nil
}
