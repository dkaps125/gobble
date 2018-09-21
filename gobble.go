package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gobble/routers"
	"gobble/utils"
)

type gobble struct {
	r chi.Router
}

func main() {
	var G gobble

	projectDir := flag.String("projectDir", "projects", "Directory in which projects will be stored")
	utils.Config.SetProjectDir(*projectDir)

	port := flag.Int("port", 3000, "Port on which the webserver will run")
	utils.Config.Port = *port

	G.init()
	G.start()
}

func (g *gobble) init() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/gitwebhook", routers.GitWebhooks)

	g.r = r
}

func (g *gobble) start() {
	if g.r == nil {
		panic("Instance not correctly initialized!")
	}

	port := fmt.Sprintf(":%d", utils.Config.Port)
	http.ListenAndServe(port, g.r)
}
