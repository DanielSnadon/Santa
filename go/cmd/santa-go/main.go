package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"santa/go/internal/app"
	"santa/go/internal/auth"
	"santa/go/internal/config"
	"santa/go/internal/engine"
	"santa/go/internal/httpapi"
	"santa/go/internal/store"
)

func main() {
	migrate := flag.Bool("migrate", false, "apply Go session migration")
	flag.Parse()
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL or ConnectionStrings__DefaultConnection is required")
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	if *migrate {
		data, err := os.ReadFile(filepath.Join("go", "migrations", "0001_go_sessions.sql"))
		if err != nil {
			data, err = os.ReadFile(filepath.Join("migrations", "0001_go_sessions.sql"))
		}
		if err != nil {
			log.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, string(data)); err != nil {
			log.Fatal(err)
		}
		return
	}
	sqlStore := &store.Store{DB: db}
	appStore := store.AppStore{Store: sqlStore}
	service := app.Service{Store: appStore, Engine: engine.Native{Budget: engine.DefaultBudget()}}
	users := auth.UserStore{DB: db}
	sessions := auth.Manager{Store: auth.SQLSessionStore{DB: db}, TTL: cfg.SessionTTL, Secure: cfg.SecureCookies}
	templates, _ := httpapi.LoadTemplates(filepath.Join("go", "templates"))
	if templates == nil {
		templates, _ = httpapi.LoadTemplates("templates")
	}
	staticDir := filepath.Join("go", "static")
	if _, err := os.Stat(staticDir); err != nil {
		staticDir = "static"
	}
	srv := &httpapi.Server{Static: http.FileServer(http.Dir(staticDir)), Auth: httpapi.SQLAuth{Users: users, Sessions: sessions}, Games: httpapi.AppAdapter{Service: service}, Rate: httpapi.NewRateLimiter(20, time.Minute), Templates: templates}
	log.Printf("santa-go listening on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
