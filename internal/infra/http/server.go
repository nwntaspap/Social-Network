package http

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arnald/forum/internal/app"
	"github.com/arnald/forum/internal/bootstrap"
	"github.com/arnald/forum/internal/config"
	"github.com/arnald/forum/internal/domain/session"
	getuseractivity "github.com/arnald/forum/internal/infra/http/activity/getUserActivity"
	"github.com/arnald/forum/internal/infra/http/authcookies"
	createcategory "github.com/arnald/forum/internal/infra/http/category/createCategory"
	deletecategory "github.com/arnald/forum/internal/infra/http/category/deleteCategory"
	getallcategories "github.com/arnald/forum/internal/infra/http/category/getAllCategories"
	getcategorybyid "github.com/arnald/forum/internal/infra/http/category/getCategoryByID"
	updatecategory "github.com/arnald/forum/internal/infra/http/category/updateCategory"
	getchatusers "github.com/arnald/forum/internal/infra/http/chat/getChatUsers"
	initchat "github.com/arnald/forum/internal/infra/http/chat/initChat"
	createcomment "github.com/arnald/forum/internal/infra/http/comment/createComment"
	deletecomment "github.com/arnald/forum/internal/infra/http/comment/deleteComment"
	getcomment "github.com/arnald/forum/internal/infra/http/comment/getComment"
	getcommentsbytopic "github.com/arnald/forum/internal/infra/http/comment/getCommentsByTopic"
	updatecomment "github.com/arnald/forum/internal/infra/http/comment/updateComment"
	"github.com/arnald/forum/internal/infra/http/health"
	getnotifications "github.com/arnald/forum/internal/infra/http/notification/getNotifications"
	getunreadcount "github.com/arnald/forum/internal/infra/http/notification/getUnreadCount"
	markallasread "github.com/arnald/forum/internal/infra/http/notification/markAllAsRead"
	markasread "github.com/arnald/forum/internal/infra/http/notification/markAsRead"
	streamnotification "github.com/arnald/forum/internal/infra/http/notification/streamNotification"
	oauthlogin "github.com/arnald/forum/internal/infra/http/oauth"
	createtopic "github.com/arnald/forum/internal/infra/http/topic/createTopic"
	deletetopic "github.com/arnald/forum/internal/infra/http/topic/deleteTopic"
	getalltopics "github.com/arnald/forum/internal/infra/http/topic/getAllTopics"
	gettopic "github.com/arnald/forum/internal/infra/http/topic/getTopic"
	updatetopic "github.com/arnald/forum/internal/infra/http/topic/updateTopic"
	getme "github.com/arnald/forum/internal/infra/http/user/getMe"
	userLogin "github.com/arnald/forum/internal/infra/http/user/login"
	"github.com/arnald/forum/internal/infra/http/user/logout"
	userRegister "github.com/arnald/forum/internal/infra/http/user/register"
	castvote "github.com/arnald/forum/internal/infra/http/vote/castVote"
	deletevote "github.com/arnald/forum/internal/infra/http/vote/deleteVote"
	getCounts "github.com/arnald/forum/internal/infra/http/vote/getVoteCounts"
	wshttp "github.com/arnald/forum/internal/infra/http/ws"
	"github.com/arnald/forum/internal/infra/logger"
	"github.com/arnald/forum/internal/infra/middleware"
	"github.com/arnald/forum/internal/infra/ws"
	wshandlers "github.com/arnald/forum/internal/infra/ws/handlers"
	oauth "github.com/arnald/forum/internal/pkg/oAuth"
)

const (
	apiContext               = "/api/v1"
	readTimeout              = 5 * time.Second
	writeTimeout             = 10 * time.Second
	idleTimeout              = 15 * time.Second
	stateManagerDefaultLimit = 10
)

type Server struct {
	appServices    app.Services
	config         *config.ServerConfig
	router         *http.ServeMux
	wsRouter       ws.WSRouter
	sessionManager session.Manager
	cookieManager  *authcookies.Manager
	oauth          *oauth.OAuth
	middleware     *middleware.Middleware
	db             *sql.DB
	logger         logger.Logger
	hub            *ws.Hub
}

func NewServer(cfg *config.ServerConfig, app *bootstrap.App) *Server {
	httpServer := &Server{
		router:         http.NewServeMux(),
		appServices:    app.Services,
		config:         cfg,
		logger:         app.Logger,
		oauth:          app.OAuth,
		cookieManager:  app.CookieManager,
		sessionManager: app.SessionManager,
		middleware:     app.Middlware,
		hub:            app.Hub,
	}
	httpServer.initWSRouter()
	httpServer.AddHTTPRoutes()
	return httpServer
}

func middlewareChain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}

func (server *Server) initWSRouter() {

	chatHistoryWShandler := wshandlers.NewChatHistoryHandler(server.appServices.Queries.GetChatHistory, server.logger)

	pingWShandler := wshandlers.NewPingHandler()
	chatOpenHandler := wshandlers.NewChatOpenHandler(server.hub)
	chatCloseHandler := wshandlers.NewChatCloseHandler(server.hub)
	chatTypingHandler := wshandlers.NewChatTypingHandler(server.hub)

	markAsReadWShandler := wshandlers.NewChatMarkReadHandler(server.appServices.Commands.MarkAsReadChatMessage, server.logger)

	sendWShandler := wshandlers.NewChatSendHandler(server.appServices.Commands.SendChatMessage, server.logger)

	server.wsRouter = ws.NewWSRouter(
		chatHistoryWShandler,
		pingWShandler,
		markAsReadWShandler,
		sendWShandler,
		chatOpenHandler,
		chatCloseHandler,
		chatTypingHandler,
	)
}
func (server *Server) AddHTTPRoutes() {
	server.router.HandleFunc(apiContext+"/health",
		middlewareChain(
			health.NewHandler(server.logger, server.appServices.Commands.CreateNotification).HealthCheck,
			server.middleware.Authorization.Required,
		))

	server.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/static"))))

	server.router.HandleFunc("/", spaHandler("frontend/static/index.html"))

	// User routes
	server.router.HandleFunc(apiContext+"/login/email",
		userLogin.NewHandler(server.config, server.appServices, server.sessionManager, server.logger, server.cookieManager).UserLoginEmail,
	)
	server.router.HandleFunc(apiContext+"/login/username",
		userLogin.NewHandler(server.config, server.appServices, server.sessionManager, server.logger, server.cookieManager).UserLoginUsername,
	)
	server.router.HandleFunc(apiContext+"/register",
		userRegister.NewHandler(server.config, server.appServices, server.sessionManager, server.logger).UserRegister,
	)
	server.router.HandleFunc(apiContext+"/logout",
		middlewareChain(
			logout.NewHandler(server.sessionManager, server.logger, server.cookieManager).Logout,
			server.middleware.Authorization.Required,
		))
	// New handler for retrieving current user data from backend
	server.router.HandleFunc(apiContext+"/me",
		middlewareChain(
			getme.NewHandler(server.logger).GetMe,
			server.middleware.Authorization.Required,
		))
	// OAuth routes
	server.router.HandleFunc(apiContext+"/auth/github/login",
		oauthlogin.NewOAuthHandler(
			server.oauth.GithubProvider,
			server.config,
			&server.appServices.Queries.UserLoginGithub,
			server.oauth.StateManager,
			server.sessionManager,
			server.logger,
		).Login,
	)
	server.router.HandleFunc(apiContext+"/auth/github/callback",
		oauthlogin.NewOAuthHandler(
			server.oauth.GithubProvider,
			server.config,
			&server.appServices.Queries.UserLoginGithub,
			server.oauth.StateManager,
			server.sessionManager,
			server.logger,
		).Callback,
	)
	server.router.HandleFunc(apiContext+"/auth/google/login",
		oauthlogin.NewOAuthHandler(
			server.oauth.GoogleProvider,
			server.config,
			&server.appServices.Queries.UserLoginGithub,
			server.oauth.StateManager,
			server.sessionManager,
			server.logger,
		).Login,
	)
	server.router.HandleFunc(apiContext+"/auth/google/callback",
		oauthlogin.NewOAuthHandler(
			server.oauth.GoogleProvider,
			server.config,
			&server.appServices.Queries.UserLoginGithub,
			server.oauth.StateManager,
			server.sessionManager,
			server.logger,
		).Callback,
	)

	// Topic routes
	server.router.HandleFunc(apiContext+"/topics/create",
		middlewareChain(
			createtopic.NewHandler(server.appServices, server.config, server.logger).CreateTopic,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/topics/update",
		middlewareChain(
			updatetopic.NewHandler(server.appServices, server.config, server.logger).UpdateTopic,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/topics/delete",
		middlewareChain(
			deletetopic.NewHandler(server.appServices, server.config, server.logger).DeleteTopic,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/topic",
		middlewareChain(
			gettopic.NewHandler(server.appServices, server.config, server.logger).GetTopic,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/topics/all",
		middlewareChain(
			getalltopics.NewHandler(server.appServices, server.config, server.logger).GetAllTopics,
			server.middleware.Authorization.Required,
		),
	)

	// Comment routes
	server.router.HandleFunc(apiContext+"/comments/create",
		middlewareChain(
			createcomment.NewHandler(server.appServices.Commands.CreateComment, server.config, server.logger).CreateComment,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/comments/update",
		middlewareChain(
			updatecomment.NewHandler(server.appServices, server.config, server.logger).UpdateComment,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/comments/delete",
		middlewareChain(
			deletecomment.NewHandler(server.appServices, server.config, server.logger).DeleteComment,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/comments/get",
		getcomment.NewHandler(server.appServices, server.config, server.logger).GetComment,
	)
	server.router.HandleFunc(apiContext+"/comments/topic",
		getcommentsbytopic.NewHandler(server.appServices, server.config, server.logger).GetCommentsByTopic,
	)

	// Category routes
	server.router.HandleFunc(apiContext+"/category/create",
		middlewareChain(
			createcategory.NewHandler(server.appServices, server.config, server.logger).CreateCategory,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/category/delete",
		middlewareChain(
			deletecategory.NewHandler(server.appServices, server.config, server.logger).DeleteCategory,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/category/update",
		middlewareChain(
			updatecategory.NewHandler(server.appServices, server.config, server.logger).UpdateCategory,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/category",
		middlewareChain(
			getcategorybyid.NewHandler(server.appServices, server.config, server.logger).GetCategoryByID,
			server.middleware.Authorization.Required,
		),
	)
	server.router.HandleFunc(apiContext+"/categories/all",
		getallcategories.NewHandler(server.appServices, server.config, server.logger).GetAllCategories,
	)

	// Vote routes
	server.router.HandleFunc(apiContext+"/vote/cast",
		middlewareChain(
			castvote.NewHandler(server.appServices.Commands.CastVote, server.config, server.logger).CastVote,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/vote/delete",
		middlewareChain(
			deletevote.NewHandler(server.appServices, server.config, server.logger).DeleteVote,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/vote/counts",
		middlewareChain(
			getCounts.NewHandler(server.appServices, server.config, server.logger).GetCounts,
			server.middleware.Authorization.Required,
		),
	)

	// Activity routes
	server.router.HandleFunc(apiContext+"/user/activity",
		middlewareChain(
			getuseractivity.NewHandler(server.appServices, server.config, server.logger).GetUserActivity,
			server.middleware.Authorization.Required,
		),
	)

	// Notifications routes

	server.router.HandleFunc(apiContext+"/notifications/stream", // get
		middlewareChain(
			streamnotification.NewHandler(server.appServices.Commands.OpenStream).StreamNotifications,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/notifications/unread-count", // get
		middlewareChain(
			getunreadcount.NewHandler(server.appServices.Queries.GetUnreadCount).GetUnread,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/notifications", // get
		middlewareChain(
			getnotifications.NewHandler(server.appServices.Queries.GetNotifications).GetNotifications,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/notifications/mark-read", // post
		middlewareChain(
			markasread.NewHandler(server.appServices.Commands.MarkAsRead).MarkAsRead,
			server.middleware.Authorization.Required,
		),
	)

	server.router.HandleFunc(apiContext+"/notifications/mark-all-read", // post
		middlewareChain(
			markallasread.NewHandler(server.appServices.Commands.MarkAllAsRead).MarkAllAsRead,
			server.middleware.Authorization.Required,
		),
	)

	// WebSocket route — chat and presence
	server.router.HandleFunc(apiContext+"/ws",
		middlewareChain(
			wshttp.NewHandler(server.hub, server.wsRouter, server.logger).UpgradeConnection,
			server.middleware.Authorization.Required,
		),
	)

	// Chat routes
	server.router.HandleFunc(apiContext+"/chat/init",
		middlewareChain(
			initchat.NewHandler(server.appServices.Commands.InitChat, server.logger).InitChat,
			server.middleware.Authorization.Required,
		))

	server.router.HandleFunc(apiContext+"/chat/users",
		middlewareChain(
			getchatusers.NewHandler(
				server.appServices.Queries.GetChatUsers,
				server.logger,
			).GetChatUsers,
			server.middleware.Authorization.Required,
		))
}

func (server *Server) ListenAndServe() {
	wrappedRouter := middleware.NewCorsMiddleware(server.router)

	if server.config.RateLimit.Enabled {
		wrappedRouter = middleware.NewRateLimiterMiddleware(
			wrappedRouter,
			server.config.RateLimit.RequestsLimit,
			server.config.RateLimit.WindowSeconds,
			server.config.RateLimit.Cleanup,
		)
		server.logger.PrintInfo("Rate Limit wrapped", nil)
		log.Printf("  2. Rate Limit middleware (limit: %d req/%ds cleanup: %s)",
			server.config.RateLimit.RequestsLimit,
			server.config.RateLimit.WindowSeconds,
			server.config.RateLimit.Cleanup.String())
	}

	srv := &http.Server{
		Addr:         server.config.Host + ":" + server.config.Port,
		Handler:      wrappedRouter,
		ReadTimeout:  server.config.ReadTimeout,
		WriteTimeout: server.config.WriteTimeout,
		IdleTimeout:  server.config.IdleTimeout,
	}
	server.logger.PrintInfo("Starting server", map[string]string{
		"host":        server.config.Host,
		"port":        server.config.Port,
		"environment": server.config.Environment,
	})

	var err error
	if server.config.TLSCertFile != "" && server.config.TLSKeyFile != "" {
		log.Printf("Starting HTTPS server with TLS certificates")
		err = srv.ListenAndServeTLS(server.config.TLSCertFile, server.config.TLSKeyFile)
	} else {
		log.Printf("Starting HTTP server (no TLS)")
		err = srv.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		server.logger.PrintFatal(err, nil)
	}
}

func spaHandler(indexPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Let API routes pass through
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Skip static routes (let the dedicated static handler process them)
		if strings.HasPrefix(r.URL.Path, "/static/") {
			http.NotFound(w, r)
			return
		}

		// For all other routes (including root and client-side routes), serve index.html
		http.ServeFile(w, r, indexPath)
	}
}
