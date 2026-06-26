package infra

import (
	"database/sql"

	"social-network/internal/bootstrap"
	"social-network/internal/config"
	"social-network/internal/infra/http"
	"social-network/internal/infra/storage/sqlite"
)

type Services struct {
	Repositories *sqlite.Repositories
	Server       *http.Server
}

func NewInfraProviders(db *sql.DB) Services {
	return Services{
		Repositories: sqlite.NewRepositories(db),
	}
}

func NewHTTPServer(cfg *config.ServerConfig, app *bootstrap.App) *http.Server {
	return http.NewServer(cfg, app)
}
