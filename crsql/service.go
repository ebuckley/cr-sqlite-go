package crsql

import (
	"connectrpc.com/connect"
	"context"
	"database/sql"
	"fmt"
	api "github.com/ebuckley/crsqlite-go/gen/api/v1"
	"log/slog"
)

type SyncService struct {
	DB     *sql.DB
	Schema string
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
	slog.Info("Get change from: ", "dbversion:", c.Msg.DbVersion)
	ch, err := GetChanges(s.DB, int(c.Msg.DbVersion))
	if err != nil {
		return nil, err
	}
	if len(ch) > 0 {
		for i, change := range ch {
			slog.Info(fmt.Sprintf("change %d: %v \t %s", i, change.String(), change.Val))
		}
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

func (s *SyncService) GetSchema(ctx context.Context, c *connect.Request[api.GetSchemaRequest]) (*connect.Response[api.GetSchemaResponse], error) {
	return connect.NewResponse(&api.GetSchemaResponse{Schema: s.Schema, Version: 0}), nil
}
