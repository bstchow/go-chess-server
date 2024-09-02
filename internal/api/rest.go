package api

import (
	"net/http"

	"github.com/bstchow/go-chess-server/pkg/logging"

	"go.uber.org/zap"
)

// Start REST server
func StartRESTServer(port string) error {
	http.HandleFunc("POST /api/login", handlerLogin)

	logging.Info("rest server started", zap.String("port", port))

	return http.ListenAndServe(":"+port, nil)
}
