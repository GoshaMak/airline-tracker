package main

import (
	"api/internal/infra"
	"api/internal/user/domain"
	userPostgres "api/internal/user/infra/postgres"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" || os.Getenv("POSTGRES_URI") == "" {
		return fmt.Errorf("ADMIN_EMAIL, ADMIN_PASSWORD and POSTGRES_URI are required")
	}
	admin, err := domain.NewUser(email, password, domain.AdminRole)
	if err != nil {
		return fmt.Errorf("invalid administrator credentials: %w", err)
	}
	injector := do.New(infra.Package)
	pool, err := do.Invoke[*pgxpool.Pool](injector)
	if err != nil {
		return err
	}
	defer pool.Close()
	repo, err := userPostgres.NewUserRepository(injector)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := repo.SaveUser(ctx, admin); err != nil {
		return fmt.Errorf("create administrator: %w", err)
	}
	fmt.Printf("Administrator created: %s\n", admin.Email.String())
	return nil
}
