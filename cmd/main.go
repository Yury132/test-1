package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Yury132/Golang-Task-1/internal/client/google"
	"github.com/Yury132/Golang-Task-1/internal/config"
	"github.com/Yury132/Golang-Task-1/internal/service"
	"github.com/Yury132/Golang-Task-1/internal/storage"
	transport "github.com/Yury132/Golang-Task-1/internal/transport/http"
	"github.com/Yury132/Golang-Task-1/internal/transport/http/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// ---------------------------новое --------------------------------
const (
	dialect  = "pgx"
	dbString = "host=localhost user=root password=mydbpass dbname=mydb port=5432 sslmode=disable"
	command  = "up"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir   = flags.String("dir", "migrations", "directory with migration files")
)

//---------------------------новое --------------------------------

func main() {

	//---------------------------новое --------------------------------
	// flags.Usage = usage
	// flags.Parse(os.Args[1:])

	// args := flags.Args()

	// if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
	// 	flags.Usage()
	// 	return
	// }

	//command := args[0]

	db, err := goose.OpenDBWithDriver(dialect, dbString)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	//if err := goose.Run(command, db, *dir, args[1:]...); err != nil {
	if err := goose.Run(command, db, *dir); err != nil {
		log.Fatalf("migrate %v: %v", command, err)
	}

	//---------------------------новое -------------------
	cfg, err := config.Parse()
	if err != nil {
		panic(err)
	}

	logger := cfg.Logger()

	poolCfg, err := cfg.PgPoolConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to db")
	}

	oauthCfg := cfg.SetupConfig()

	googleAPI := google.New(logger)

	strg := storage.New(conn)
	svc := service.New(logger, oauthCfg, googleAPI, strg)
	handler := handlers.New(logger, oauthCfg, svc)
	srv := transport.New(":8080").WithHandler(handler)

	// graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT)

	go func() {
		if err = srv.Run(); err != nil {
			logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-shutdown
}

// ---------------------------новое -------------------
// func usage() {
// 	fmt.Println(usagePrefix)
// 	flags.PrintDefaults()
// 	fmt.Println(usageCommands)
// }

// var (
// 	usagePrefix = `Usage: migrate COMMAND
// Examples:
//     migrate status
// `

// 	usageCommands = `
// Commands:
//     up                   Migrate the DB to the most recent version available
//     up-by-one            Migrate the DB up by 1
//     up-to VERSION        Migrate the DB to a specific VERSION
//     down                 Roll back the version by 1
//     down-to VERSION      Roll back to a specific VERSION
//     redo                 Re-run the latest migration
//     reset                Roll back all migrations
//     status               Dump the migration status for the current DB
//     version              Print the current version of the database
//     create NAME [sql|go] Creates new migration file with the current timestamp
//     fix                  Apply sequential ordering to migrations`
// )

//---------------------------новое --------------------------------
