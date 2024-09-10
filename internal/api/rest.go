package api

import (
	"net/http"

	"github.com/bstchow/go-chess-server/pkg/agent"
	"github.com/bstchow/go-chess-server/pkg/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// Start REST server
func StartRESTServer(port string, agent *agent.Agent) error {

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// TODO: restrict cors to only allow requests from an ENV variable set of domains
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Post("/api/privyLogin", handlerPrivyLogin)
	r.Post("/api/fcFrameLogin", handlerFcFrameLogin)
	r.Get("/api/sessionCount", injectHandlerSessionCount(agent))
	logging.Info("rest server started", zap.String("port", port))

	return http.ListenAndServe(":"+port, r)
}
