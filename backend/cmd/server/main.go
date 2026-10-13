package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidyoung/savers-app/backend/internal/db"
	"github.com/davidyoung/savers-app/backend/internal/handlers"
	mw "github.com/davidyoung/savers-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	addr := env("ADDR", ":8080")
	dbPath := env("DB_PATH", "./data/savers.db")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer database.Close()

	authH := &handlers.AuthHandler{DB: database}
	routineH := &handlers.RoutineHandler{DB: database}
	sessionH := &handlers.SessionHandler{DB: database}
	contentH := &handlers.ContentHandler{DB: database}
	journalH := &handlers.JournalHandler{DB: database}
	statsH := &handlers.StatsHandler{DB: database}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Timezone"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"savers"}`))
	})

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/register", authH.Register)
		api.Post("/auth/login", authH.Login)

		api.Group(func(priv chi.Router) {
			priv.Use(mw.Auth)
			priv.Get("/auth/me", authH.Me)

			priv.Get("/routine", routineH.Get)
			priv.Put("/routine", routineH.Update)

			priv.Get("/sessions/today", sessionH.Today)
			priv.Post("/sessions/start", sessionH.Start)
			priv.Patch("/sessions/{id}/steps/{key}", sessionH.UpdateStep)
			priv.Post("/sessions/{id}/complete", sessionH.Complete)

			priv.Get("/content", contentH.List)
			priv.Get("/content/{id}", contentH.Get)

			priv.Get("/journals", journalH.List)
			priv.Post("/journals", journalH.Create)
			priv.Put("/journals/{id}", journalH.Update)

			priv.Get("/stats", statsH.Stats)
			priv.Get("/calendar", statsH.Calendar)
		})
	})

	staticDir := env("STATIC_DIR", "")
	if staticDir != "" {
		abs, _ := filepath.Abs(staticDir)
		fileServer := http.FileServer(http.Dir(abs))
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/api") {
				http.NotFound(w, req)
				return
			}
			path := filepath.Join(abs, filepath.Clean(req.URL.Path))
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, req)
				return
			}
			http.ServeFile(w, req, filepath.Join(abs, "index.html"))
		})
	}

	log.Printf("SAVERS listening on %s (db=%s static=%s)", addr, dbPath, staticDir)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
