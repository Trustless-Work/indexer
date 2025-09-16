package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/Trustless-Work/indexer/internal/db"
	// "github.com/Trustless-Work/indexer/internal/db/migrate"
	"github.com/Trustless-Work/indexer/internal/deposits"
	"github.com/Trustless-Work/indexer/internal/escrow"
	"github.com/Trustless-Work/indexer/internal/httpserver"
	"github.com/Trustless-Work/indexer/internal/rpc"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB pool
	pool, err := db.NewPool(ctx, db.ConfigFromEnv())
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	dsn := os.Getenv("DB_DSN")
	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("DB_DSN=%s", dsn) // debe verse :15432/trustlesswork
	// Skip migrations for now since DB is already restored from dump
	// if err := migrate.Up(ctx, dsn); err != nil {
	//	 log.Fatalf("migrations failed: %v", err)
	// }

	// RPC client
	var rpcClient rpc.Client
	if os.Getenv("RPC_USE_MOCK") == "true" {
		rpcClient = rpc.NewMockClient()
	} else {
		rpcClient = rpc.NewHTTPClient(os.Getenv("SOROBAN_RPC_URL")) // TODO: implementar real
	}

	// Repositories (SQL fallback; cambiaremos a SP cuando me pases firmas)
	singleRepo := escrow.NewSingleSQLRepository(pool)
	multiRepo := escrow.NewMultiSQLRepository(pool)
	depRepo := deposits.NewSQLRepository(pool)

	// Services
	escrowSvc := escrow.NewService(singleRepo, multiRepo)
	depSvc := deposits.NewService(depRepo, rpcClient)

	// HTTP server
	srv := httpserver.New(os.Getenv("HTTP_ADDR"), escrowSvc, depSvc)

	errCh := make(chan error, 1)
	go func() {
		// log.Printf("HTTP listening on %s", srv.Addr())
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("shutting down...")
		shCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shCtx)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}
}
