package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/ghazlabs/wa-scheduler/internal/core"
	wa "github.com/ghazlabs/wa-scheduler/internal/driven/publisher"
	"github.com/ghazlabs/wa-scheduler/internal/driven/scheduler"
	"github.com/ghazlabs/wa-scheduler/internal/driven/storage"
	"github.com/ghazlabs/wa-scheduler/internal/driver"
	"github.com/go-co-op/gocron/v2"
	"github.com/go-resty/resty/v2"

	_ "modernc.org/sqlite"
)

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	waPublisher, err := wa.NewWaPublisher(wa.WaPublisherConfig{
		HttpClient:   resty.New(),
		Username:     cfg.WAPublisherUsername,
		Password:     cfg.WAPublisherPassword,
		WaApiBaseUrl: cfg.WAPublisherApiBaseUrl,
	})
	if err != nil {
		log.Fatalf("failed to create wa publisher: %v", err)
	}

	gocronClient, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("failed to create gocron client: %v", err)
	}
	gocronClient.Start()
	defer func() {
		if err := gocronClient.Shutdown(); err != nil {
			log.Fatalf("failed to stop gocron client: %v", err)
		}
	}()

	dbClient, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize sqlite client: %v", err)
	}

	messageStorage, err := storage.NewStorage(storage.StorageConfig{
		DB: dbClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	gocronScheduler, err := scheduler.NewGoCronScheduler(scheduler.GoCronSchedulerConfig{
		Client:    gocronClient,
		Publisher: waPublisher,
		Storage:   messageStorage,
	})
	if err != nil {
		log.Fatalf("failed to create gocron scheduler: %v", err)
	}

	service, err := core.NewService(core.ServiceConfig{
		Storage:   messageStorage,
		Scheduler: gocronScheduler,
	})
	if err != nil {
		log.Fatalf("failed to create service: %v", err)
	}
	service.InitializeService(context.Background())

	api, err := driver.NewAPI(driver.APIConfig{
		Service:            service,
		DefaultNumbers:     cfg.WADefaultNumbers,
		ClientUsername:     cfg.ClientUsername,
		ClientPassword:     cfg.ClientPassword,
		WebClientPublicDir: cfg.WebClientPublicDir,
	})
	if err != nil {
		log.Fatalf("failed to create api: %v", err)
	}

	// initialize server
	listenAddr := fmt.Sprintf(":%s", cfg.ListenPort)
	s := &http.Server{
		Addr:        listenAddr,
		Handler:     api.GetHandler(),
		ReadTimeout: time.Second * 30,
	}
	// run server
	log.Printf("server is listening on %v", listenAddr)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatalf("unable to run server due: %v", err)
	}
}

type config struct {
	ListenPort string `env:"LISTEN_PORT,required" envDefault:"9866"`

	DBPath string `env:"DB_PATH,required" envDefault:"/data/wa-scheduler.db"`

	ClientUsername string `env:"DASHBOARD_CLIENT_USERNAME,required"`
	ClientPassword string `env:"DASHBOARD_CLIENT_PASSWORD,required"`

	WADefaultNumbers      []string `env:"WA_DEFAULT_NUMBERS"`
	WAPublisherApiBaseUrl string   `env:"WA_PUBLISHER_API_BASE_URL,required"`
	WAPublisherUsername   string   `env:"WA_PUBLISHER_USERNAME,required"`
	WAPublisherPassword   string   `env:"WA_PUBLISHER_PASSWORD,required"`

	WebClientPublicDir string `env:"WEB_CLIENT_PUBLIC_DIR,required" envDefault:"web"`
}
