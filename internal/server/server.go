package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"hyttekos/internal/akiles"
	"hyttekos/internal/config"
	"hyttekos/internal/handlers"
	"hyttekos/internal/middleware"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

type Server struct {
	config           *config.Config
	db               *sql.DB
	akilesClient     *akiles.Client
	sessionStore     sessions.Store
	tempRepo         *repository.TemperatureRepository
	sessionRepo      *repository.SessionRepository
	engine           *gin.Engine
}

func New(cfg *config.Config, db *sql.DB) *Server {
	// Create repositories
	tempRepo := repository.NewTemperatureRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Create Akiles client
	akilesClient := akiles.NewClient(
		cfg.AkilesClientID,
		cfg.AkilesClientSecret,
		cfg.AkilesHeatingGadgetID,
		cfg.AkilesHotWaterGadgetID,
	)

	// Create session store
	fmt.Printf("Server: Creating session store with secret length: %d\n", len(cfg.SessionSecret))
	if cfg.SessionSecret == "" {
		fmt.Printf("Server: ERROR - Session secret is empty!\n")
	}
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = true // Enable secure cookies for HTTPS
	sessionStore.Options.SameSite = http.SameSiteLaxMode
	sessionStore.Options.Path = "/"
	sessionStore.Options.MaxAge = 86400 * 30 // 30 days
	fmt.Printf("Server: Session store created with options: HttpOnly=%v, Secure=%v, MaxAge=%d\n", 
		sessionStore.Options.HttpOnly, sessionStore.Options.Secure, sessionStore.Options.MaxAge)

	server := &Server{
		config:       cfg,
		db:           db,
		akilesClient: akilesClient,
		sessionStore: sessionStore,
		tempRepo:     tempRepo,
		sessionRepo:  sessionRepo,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)
	
	s.engine = gin.New()
	s.engine.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
	s.engine.Use(middleware.CORS())

	// Security middleware
	s.engine.Use(middleware.Security())

	// Load HTML templates
	s.engine.LoadHTMLGlob("web/templates/*")
	s.engine.Static("/static", "web/static")

	// Create handlers
	authHandler := handlers.NewAuthHandler(s.akilesClient, s.sessionStore, s.sessionRepo, s.config.OAuthRedirectURL)
	apiHandler := handlers.NewAPIHandler(s.akilesClient, s.tempRepo, s.sessionRepo)
	webHandler := handlers.NewWebHandler(s.akilesClient, s.tempRepo, s.sessionRepo, s.sessionStore)
	tempHandler := handlers.NewTemperatureHandler(s.tempRepo)

	// Authentication routes
	auth := s.engine.Group("/auth")
	{
		auth.GET("/login", authHandler.Login)
		auth.GET("/callback", authHandler.Callback)
		auth.POST("/logout", middleware.RequireAuth(s.sessionStore, s.sessionRepo, s.akilesClient), authHandler.Logout)
	}

	// API routes (JSON)
	api := s.engine.Group("/api")
	api.Use(middleware.RequireAuth(s.sessionStore, s.sessionRepo, s.akilesClient))
	{
		api.GET("/status", apiHandler.GetStatus)
		api.POST("/heating", apiHandler.SetHeating)
		api.POST("/hot-water", apiHandler.SetHotWater)
		api.GET("/temperature", apiHandler.GetTemperature)
		api.GET("/temperature/history", apiHandler.GetTemperatureHistory)
	}

	// Form toggle routes
	toggle := s.engine.Group("/toggle")
	toggle.Use(middleware.RequireAuth(s.sessionStore, s.sessionRepo, s.akilesClient))
	{
		toggle.POST("/heating", webHandler.ToggleHeating)
		toggle.POST("/hot-water", webHandler.ToggleHotWater)
	}

	// Temperature sensor endpoint (no auth required)
	s.engine.POST("/api/temperature", tempHandler.SubmitTemperature)

	// Main web routes
	s.engine.GET("/", middleware.RequireAuth(s.sessionStore, s.sessionRepo, s.akilesClient), webHandler.Index)
	s.engine.GET("/login", webHandler.LoginPage)

	// Health check
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return s.engine.Run(addr)
}