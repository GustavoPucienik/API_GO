package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"

	"systemapi/internal/admin"
	"systemapi/internal/db"
	"systemapi/internal/fiscal"
	"systemapi/internal/logistic"
	"systemapi/internal/maintenance"
	"systemapi/pkg/config"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	log.Println("database connected")

	queries := db.New(sqlDB)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors)

	adminHandler := admin.NewHandler(sqlDB, queries)
	adminHandler.RegisterRoutes(r, sqlDB)

	maintenanceHandler := maintenance.NewHandler(sqlDB, queries)
	maintenanceHandler.RegisterRoutes(r, sqlDB)

	fiscalHandler := fiscal.NewHandler(sqlDB, queries)
	fiscalHandler.RegisterRoutes(r, sqlDB)

	logisticHandler := logistic.NewHandler(sqlDB, queries)
	logisticHandler.RegisterRoutes(r, sqlDB)

	addr := ":" + cfg.Port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
