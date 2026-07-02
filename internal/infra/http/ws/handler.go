package ws

import (
	"net/http"
	"slices"

	"social-network/internal/infra/logger"
	"social-network/internal/infra/middleware"
	"social-network/internal/infra/ws"
	"social-network/internal/pkg/helpers"

	"github.com/gorilla/websocket"
)

type Handler struct {
	hub            *ws.Hub
	router         ws.WSRouter
	logger         logger.Logger
	upgrader       websocket.Upgrader
	allowedOrigins []string
}

func NewHandler(hub *ws.Hub, router ws.WSRouter, logger logger.Logger, allowedOrigins []string) *Handler {
	return &Handler{
		hub:            hub,
		router:         router,
		logger:         logger,
		allowedOrigins: allowedOrigins,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return slices.Contains(allowedOrigins, origin)
			},
		},
	}
}

func (h *Handler) UpgradeConnection(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.PrintError(err, nil)
		return
	}

	client := ws.NewClient(user.ID, h.hub, conn)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(h.router.Route)
}
