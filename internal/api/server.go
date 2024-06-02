package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/api/middleware"
)

type Server interface {
	// Run runs the server.
	Run(ctx context.Context) error
	// Mount adds the provided route groups to the server.
	// If the server is not initialized, it will add a health check route.
	Mount(routes ...RouteGroup) error
	// Shutdown gracefully shuts down the server.
	Shutdown(ctx context.Context) error
}

// RouteGroup is a route to register a sub-app to.
type RouteGroup struct {
	// Path is the path of the route.
	Path string
	// App is the fiber sub-app to use.
	App fiber.Router
}

type Config struct {
	// Address is the address to listen on.
	Address string `yaml:"address" mapstructure:"address"`
}

type server struct {
	config      *Config
	initialized bool
	app         *fiber.App
	router      fiber.Router
	mu          sync.Mutex
}

func NewServer(c *Config) Server {
	app := fiber.New()
	return &server{
		config:      c,
		initialized: false,
		app:         app,
		router:      app.Group("/api"),
	}
}

func (s *server) Run(ctx context.Context) error {
	_ = s.router.Use(middleware.Context(ctx))
	if !s.initialized {
		err := s.Mount()
		if err != nil {
			return err
		}
	}

	return s.app.Listen(s.config.Address)
}

func (s *server) Mount(routes ...RouteGroup) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.app.Server() != nil {
		return &ErrAlreadyRunning{}
	}

	if !s.initialized {
		routes = append(routes, RouteGroup{
			Path: "/",
			App: fiber.New().Get("/healthz", func(c fiber.Ctx) error {
				logger.FromContext(c.UserContext()).DebugContext(c.Context(), "healthz")
				return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
			}),
		})
		s.initialized = true
	}

	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = rErr
				return
			}
			err = fmt.Errorf("failed to mount routes: %v", r)
		}
	}()

	for _, r := range routes {
		s.router.Use(r.Path, r.App)
	}

	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
