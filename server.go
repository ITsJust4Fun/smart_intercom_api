package main

import (
	"log"
	"net/http"
	"os"
	"smart_intercom_api/graph"
	"smart_intercom_api/graph/generated"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/pkg/config"
	"smart_intercom_api/pkg/subscriptions"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config.ReadConfigFile()
	subscriptions.Init()

	router := chi.NewRouter()
	router.Use(auth.Middleware())
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	router.Handle("/playground", playground.Handler("GraphQL playground", "/api"))
	router.Handle("/api", srv)

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
