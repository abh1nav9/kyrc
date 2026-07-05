// Command kyrc-leaderboard is the API in front of Neon/Postgres for the
// kyrc leaderboard. The CLI cannot talk to Postgres directly (shipping DB
// credentials in a public binary would let anyone bypass every check and
// write straight to the table), so this service is the only thing holding
// the DATABASE_URL. It verifies Ed25519 signatures and REPLAYS keystroke
// logs before storing anything.
//
// Config (all via env, never hardcoded):
//
//	DATABASE_URL   Neon/Postgres connection string (required)
//	PORT           listen port (default 8080)
//	CORS_ORIGIN    allowed browser origin for the website (default *)
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required (never hardcode it; set it in the environment)")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	srv := &server{db: pool, cors: getenv("CORS_ORIGIN", "*")}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", srv.health)
	mux.HandleFunc("/register", srv.handleRegister)
	mux.HandleFunc("/submit", srv.handleSubmit)
	mux.HandleFunc("/leaderboard", srv.handleLeaderboard)

	h := &http.Server{
		Addr:         ":" + port,
		Handler:      srv.withCORS(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Printf("kyrc-leaderboard listening on :%s", port)
	log.Fatal(h.ListenAndServe())
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
