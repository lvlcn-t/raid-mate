package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/api/middleware"
)

type Server interface {
	// Run runs the server.
	// Runs indefinitely until an error occurs or the server is shut down.
	//
	// If no routes were mounted before, it will mount a health check route.
	//
	// Example setup:
	//
	//	srv := api.NewServer(&Config{Address: ":8080"})
	//	err := srv.Mount(RouteGroup{
	//		Path: "/v1",
	//		App: fiber.New().Get("/hello", func(c fiber.Ctx) error {
	//			return c.SendString("Hello, World!")
	//		}),
	//	})
	//	if err != nil {
	//		// handle error
	//	}
	//
	//	_ = srv.Run(context.Background())
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

func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("api.address is required")
	}
	return nil
}

type server struct {
	config      *Config
	initialized bool
	running     bool
	app         *fiber.App
	router      fiber.Router
	mu          sync.Mutex
}

// NewServer creates a new server with the provided configuration.
func NewServer(c *Config) Server {
	app := fiber.New()
	return &server{
		config:      c,
		initialized: false,
		app:         app,
		router:      app.Group("/api"),
	}
}

// Run runs the server.
// Runs indefinitely until an error occurs or the server is shut down.
func (s *server) Run(ctx context.Context) error {
	_ = s.router.Use(middleware.Context(ctx), middleware.Recover(), middleware.Logger())
	if !s.initialized {
		err := s.Mount()
		if err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.running = true
	s.mu.Unlock()
	return s.app.Listen(s.config.Address)
}

// Mount adds the provided route groups to the server.
// If the server is not initialized, it will add a health check route.
func (s *server) Mount(routes ...RouteGroup) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return &ErrAlreadyRunning{}
	}

	if !s.initialized {
		_ = s.app.Get("/healthz", func(c fiber.Ctx) error {
			logger.FromContext(c.UserContext()).DebugContext(c.Context(), "Health check")
			return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok"})
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

// Params returns the parameter with the given name from the context and converts it to the given type.
func Params[T any](ctx fiber.Ctx, name string, parse func(string) (T, error)) (T, error) {
	var empty T
	v := ctx.Params(name)
	if v == "" {
		return empty, fmt.Errorf("missing parameter %q", name)
	}
	if parse == nil {
		if _, ok := any(empty).(string); ok {
			return any(v).(T), nil
		}
		return empty, fmt.Errorf("no parser provided for parameter %q", name)
	}

	return parse(v)
}

// BadRequestResponse sends a bad request response.
func BadRequestResponse(ctx fiber.Ctx, msg string) error {
	return ctx.Status(http.StatusBadRequest).JSON(NewErrorResponse(msg, http.StatusBadRequest))
}

// InternalServerErrorResponse sends an internal server error response.
func InternalServerErrorResponse(ctx fiber.Ctx, msg string) error {
	return ctx.Status(http.StatusInternalServerError).JSON(NewErrorResponse(msg, http.StatusInternalServerError))
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(msg string, status int) fiber.Map {
	code := http.StatusText(status)
	if code == "" {
		code = "Unknown"
	}

	return fiber.Map{"error": fiber.Map{"message": msg, "code": code}}
}
