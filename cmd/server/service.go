package main

import (
	"github.com/ebuckley/crsqlite-go/crsql"
	"sync"
)

var schema = `
create table if not exists note (id primary key not null, title, body);
select crsql_as_crr('note');
`
var setup = sync.OnceFunc(func() {
	crsql.Register("/home/ersin/Code/crdt-sql/crsql/crsqlite") // TODO make configurable
})

func newSyncService() (*crsql.SyncService, error) {
	setup()
	db, err := crsql.New(":memory:", schema)
	if err != nil {
		return nil, err
	}
	return &crsql.SyncService{DB: db, Schema: schema}, nil
}
