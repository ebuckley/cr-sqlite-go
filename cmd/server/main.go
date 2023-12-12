package main

import (
	"flag"
	"fmt"
	"github.com/ebuckley/crsqlite-go/gen/api/v1/api_v1connect"
	"log/slog"
	"net/http"
	"os"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	listenOn := fmt.Sprintf("localhost:%d", *port)

	s, err := newSyncService()
	if err != nil {
		slog.Error("failed to create sync service:", "err", err)
		os.Exit(1)
	}
	path, handler := api_v1connect.NewChangeServiceHandler(s)
	slog.Info("Registered ChangeServiceHandler", "path", path)

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	slog.Info("Started server on", "port", *port)

	err = http.ListenAndServe(listenOn, cors(mux))
	if err != nil {
		slog.Error("failed to serve:", "err", err)
		os.Exit(1)
	}
}
