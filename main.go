package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/therahulprasad/spiders/crawler"
)

const (
	VERSION     = "0.5"
	MAJORCHANGE = "Batch Processing using batch_file"
)

type App struct {
	configPath string
	resume     bool
	exitChan   chan bool
	killChan   chan bool
	quitChan   chan os.Signal
}

func NewApp() *App {
	return &App{
		exitChan: make(chan bool),
		killChan: make(chan bool),
		quitChan: make(chan os.Signal, 2),
	}
}

func (app *App) parseFlags() {
	var (
		showVersion = flag.Bool("v", false, "Show version information")
		configPath  = flag.String("c", "", "Path to config.yaml file (required)")
		resume      = flag.Bool("r", false, "Resume incomplete project")
	)
	flag.Parse()

	if *showVersion {
		app.printVersion()
		os.Exit(0)
	}

	if *configPath == "" {
		log.Fatal("Config file path is required. Use -c flag to specify config.yaml")
	}

	app.configPath = *configPath
	app.resume = *resume
}

func (app *App) printVersion() {
	fmt.Printf("Spiders Web Crawler v%s\n", VERSION)
	fmt.Printf("Latest: %s\n", MAJORCHANGE)
}

func (app *App) setupSignalHandling() {
	signal.Notify(app.quitChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-app.quitChan
		fmt.Println("\nShutdown signal received. Stopping crawler...")
		app.killChan <- true
	}()
}

func (app *App) run() {
	app.parseFlags()
	app.setupSignalHandling()

	fmt.Printf("Starting Spiders crawler with config: %s\n", app.configPath)
	if app.resume {
		fmt.Println("Resuming previous crawl session")
	}

	crawler.Initialize(app.configPath, app.exitChan, app.killChan, app.resume)

	<-app.exitChan
	fmt.Println("Crawler stopped successfully. Goodbye!")
}

func main() {
	app := NewApp()
	app.run()
}
