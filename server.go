package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"log"
	"net/http"
	"os"
	"smart_intercom_api/graph"
	"smart_intercom_api/graph/generated"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/internal/plugin"
	"smart_intercom_api/pkg/config"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config.ReadConfigFile()

	router := chi.NewRouter()
	router.Use(auth.Middleware())
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	router.Handle("/playground", playground.Handler("GraphQL playground", "/api"))
	router.Handle("/api", srv)

	router.Route("/plugin", func(r chi.Router) {
		r.Get("/auth", plugin.RegisterPlugin)
		r.Get("/get_event", plugin.GetEvent)
		r.Get("/incoming_call", plugin.IncomingCall)
		r.Get("/rejected_call", plugin.RejectedCall)
		r.Get("/answer", plugin.Answer)
		r.Get("/cancel", plugin.Cancel)
		r.Get("/intercom_command", plugin.IntercomCommand)
		r.Get("/open", plugin.Open)
		r.Get("/reject", plugin.Reject)
	})

	log.Printf("connect to http://localhost:%s/playground for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
