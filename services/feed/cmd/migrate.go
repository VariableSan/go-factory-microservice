//go:build migrate

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/VariableSan/go-factory-microservice/pkg/common/config"
	"github.com/VariableSan/go-factory-microservice/pkg/common/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		direction = flag.String("direction", "", "Migration direction: up, down, drop")
		steps     = flag.Int("steps", 0, "Number of migration steps (for up/down)")
		version   = flag.Uint("version", 0, "Target version (for goto)")
	)
	flag.Parse()

	if *direction == "" {
		fmt.Println("Usage: go run -tags migrate cmd/migrate.go -direction=<up|down|drop|goto|force> [-steps=N] [-version=N] [-force=N]")
		os.Exit(1)
	}

	// Load feed configuration
	feedCfg := config.LoadFeedConfig()

	// Initialize logger
	log := logger.NewLogger(logger.Config{
		ServiceName: "feed-migrate",
		Level:       "info",
		Format:      "text",
	})

	if feedCfg.DatabaseURL == "" {
		log.Error("FEED_DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	// Initialize migrator
	m, err := migrate.New(
		"file://migrations",
		feedCfg.DatabaseURL,
	)
	if err != nil {
		log.Error("Failed to create migrator", "error", err)
		os.Exit(1)
	}
	defer m.Close()

	// Execute migration based on direction
	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
			log.Info("Running migration up", "steps", *steps)
		} else {
			err = m.Up()
			log.Info("Running migration up to latest")
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
			log.Info("Running migration down", "steps", *steps)
		} else {
			err = m.Down()
			log.Info("Running migration down (all)")
		}
	case "drop":
		err = m.Drop()
		log.Info("Dropping all tables")
	case "goto":
		if *version == 0 {
			log.Error("Version is required for goto command")
			os.Exit(1)
		}
		err = m.Migrate(*version)
		log.Info("Migrating to version", "version", *version)
	case "force":
		if *version == 0 {
			log.Error("Version is required for force command")
			os.Exit(1)
		}
		err = m.Force(int(*version))
		log.Info("Forcing version", "version", *version)
	default:
		log.Error("Invalid direction", "direction", *direction)
		os.Exit(1)
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("No migrations to run")
		} else {
			log.Error("Migration failed", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Migration completed successfully")
	}
}
