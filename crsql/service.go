package crsql

import (
	"connectrpc.com/connect"
	"context"
	"database/sql"
	api "github.com/ebuckley/crsqlite-go/gen/api/v1"
)

type SyncService struct {
	DB *sql.DB
}

func (s *SyncService) GetSiteID(ctx context.Context, c *connect.Request[api.GetSiteIDRequest]) (*connect.Response[api.GetSiteIDResponse], error) {
	id, err := GetSiteID(s.DB)
	if err != nil {
		return nil, err
	}
	r := &api.GetSiteIDResponse{SiteId: id.String()}
	return connect.NewResponse(r), nil
}

func (s *SyncService) GetChanges(ctx context.Context, c *connect.Request[api.GetChangesRequest]) (*connect.Response[api.GetChangesResponse], error) {
	ch, err := GetChanges(s.DB, int(c.Msg.DbVersion))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&api.GetChangesResponse{Changes: ch}), nil
}

func (s *SyncService) MergeChanges(ctx context.Context, c *connect.Request[api.MergeChangesRequest]) (*connect.Response[api.MergeChangesResponse], error) {
	err := MergeChanges(s.DB, c.Msg.Changes)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&api.MergeChangesResponse{}), nil
}
