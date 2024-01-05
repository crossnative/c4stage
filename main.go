package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sethvargo/go-retry"
	"github.io/remast/c4stage/backstage"
	"github.io/remast/c4stage/c4"
	"github.io/remast/c4stage/catalog"
	"github.io/remast/c4stage/shared"
)

func main() {
	log.Println("Welcome to C4 Stage - Let's go!")
	config, driver, router, err := newApp()
	if err != nil {
		log.Printf("%s\n", err)
		os.Exit(1)
	}

	defer driver.Close(context.Background())

	err = http.ListenAndServe(":"+config.BindPort, router)
	if err != nil {
		log.Printf("%s\n", err)
		os.Exit(1)
	}
}

func newApp() (*shared.Config, neo4j.DriverWithContext, *chi.Mux, error) {
	var config shared.Config
	err := envconfig.Process("c4stage", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	dbUri := config.Db
	dbUser := config.DbUser
	dbPassword := config.DbPassword
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithMaxDuration(1*time.Minute, b)
	err = retry.Do(ctx, b, func(ctx context.Context) error {
		err = driver.VerifyConnectivity(ctx)
		if err != nil {
			log.Printf("Could not verify db connection to %v, trying again", dbUri)
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	c4Repository := &c4.C4EntityNeo4j{
		Driver: driver,
	}
	catalogRepository := &catalog.CatalogRepositoryNeo4j{
		Driver: driver,
	}

	err = catalogRepository.Reset(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Importing Backstage Catalog in %v seconds.", config.BackstageImportDelay)
	if config.BackstageImportDelay != -1 {
		time.AfterFunc(time.Duration(config.BackstageImportDelay)*time.Second, func() {
			log.Println("Importing Backstage Catalog ...")
			backstageImportService := backstage.BackstageImporter{
				Config:     &config,
				Repository: catalogRepository,
			}

			err = backstageImportService.ImportBackstageCatalog()
			// err = backstageImportService.ImportYamlFiles()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Successfully imported Backstage Catalog.")
		})
	}

	apiHandlers := []shared.DomainHandler{
		&catalog.CatalogController{
			Config:     &config,
			Repository: catalogRepository,
		},
		&c4.C4Controller{
			Config:     &config,
			Repository: c4Repository,
		},
		&shared.VersionController{},
	}

	router := chi.NewRouter()
	registerRoutes(router, apiHandlers)

	return &config, driver, router, nil
}

func registerRoutes(router *chi.Mux, apiHandlers []shared.DomainHandler) {
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	router.Mount("/api", registerApiRoutes(apiHandlers))
}

func registerApiRoutes(apiHandlers []shared.DomainHandler) http.Handler {
	router := chi.NewRouter()

	for _, apiHandler := range apiHandlers {
		apiHandler.RegisterOpen(router)
	}

	router.Group(func(r chi.Router) {
		for _, apiHandler := range apiHandlers {
			apiHandler.RegisterProtected(r)
		}
	})

	return router
}
