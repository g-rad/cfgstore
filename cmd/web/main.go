package main

import (
	"github.com/g-rad/cfgstore"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"net/http"
	"os"
)

func main() {
	s := &Server{}
	if err := s.Init(); err != nil {
		panic(err)
	}
	s.InitRouting()
	s.Start()
}

type Server struct {
	Config *Config
	Repo   *cfgstore.Repository

	echoServer *echo.Echo
}

func (s *Server) Init() error {
	cfg, err := InitConfig()
	if err != nil {
		return errors.Wrap(err, "error initializing server")
	}
	s.Config = cfg

	repo, err := cfgstore.NewRepository(cfg.DbConnection)
	if err != nil {
		return errors.Wrap(err, "error initializing repository")
	}
	s.Repo = repo

	return nil
}

func (s *Server) Start() {
	s.echoServer.Logger.Fatal(s.echoServer.Start(s.Config.Address))
}

func (s *Server) InitRouting() {
	e := echo.New()

	e.Use(middleware.Logger())

	e.GET("/api/v1/config/:key", s.handleConfigGet())
	e.GET("/ping", s.handlePing())

	s.echoServer = e
}

func (s *Server) handleConfigGet() echo.HandlerFunc {

	type keyValue struct {
		Key string `json:"key"`
		Value string `json:"value"`
	}

	return func(c echo.Context) error {

		key := c.Param("key")

		values, err := s.Repo.ConfigGet(key)
		if err != nil {
			return err
		}

		resp := make([]*keyValue, len(values))
		for i, v := range values {
			resp[i] = &keyValue{Key: v.Key, Value: v.Value}
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func (s *Server) handlePing() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}
}

type Config struct {
	Address      string
	DbConnection string
}

func InitConfig() (*Config, error) {
	return &Config{
		Address:      os.Getenv("address"),
		DbConnection: os.Getenv("dbconnection"),
	}, nil
}
