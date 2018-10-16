package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"gobble/deploy"
	"gobble/routers"
	"gobble/utils"
)

type gobble struct {
	r chi.Router
}

func handleInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				log.Println("Shutting down...")
				deploy.Shutdown()
				os.Exit(0)
			}
		}
	}()
}

func main() {
	var G gobble

	projectDir := flag.String("projectDir", "projects", "Directory in which projects will be stored")
	port := flag.Int("port", 3000, "Port on which the webserver will run")
	archiveDir := flag.String("archiveDir", "archives", "Directory in which tarred projects will be stored")
	timeout := flag.Int("timeout", 30, "Timeout for build and test tasks")
	secret := flag.String("secret", "", "Global secret for webhooks")
	noDocker := flag.Bool("nodocker", false, "Whether or not to use docker")
	flag.Parse()

	absProj, err := filepath.Abs(*projectDir)
	if err != nil {
		log.Fatalln("Project directory path could not be set")
	}

	absArch, err := filepath.Abs(*archiveDir)
	if err != nil {
		log.Fatalln("Archive directory path could not be set")
	}

	utils.Config.SetProjectDir(absProj)
	utils.Config.Port = *port
	utils.Config.SetArchiveDir(absArch)
	utils.Config.Timeout = *timeout
	utils.Config.Secret = []byte(*secret)
	utils.Config.NoDocker = *noDocker

	utils.Config.WorkingDir, _ = os.Getwd()

	absPid, err := filepath.Abs("pid")
	if err != nil {
		log.Fatalln("PID directory path could not be set")
	}
	utils.Config.SetPidDir(absPid)

	handleInterrupt()

	G.initRouter()
	G.start()
}

func (g *gobble) initRouter() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/", routers.Routes())

	g.r = r

	if !utils.Config.NoDocker {
		deploy.InitDocker()
	}
}

func (g *gobble) start() {
	if g.r == nil {
		log.Panicln("Instance not correctly initialized!")
	}

	port := fmt.Sprintf(":%d", utils.Config.Port)
	http.ListenAndServe(port, g.r)
}
