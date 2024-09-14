package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kasante1/go-api/internal/data"
	"github.com/kasante1/go-api/internal/jsonlog"
	"github.com/kasante1/go-api/internal/mailer"
	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	 "strconv"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps float64
		burst int
		enabled bool
	}
	smtp struct {
		host string
		port int
		username string
		password string
		sender string
		}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	
}

func main() {
	var cfg config
	err := godotenv.Load()
	if err != nil {
	  log.Fatal("Error loading .env file")
	}

	MOVIE_DB_DSN := os.Getenv("MOVIE_DB_DSN")
	SMTP_HOST := os.Getenv("SMTP_HOST")
	SMTP_PORT_STRING := os.Getenv("SMTP_PORT")
	SMTP_PORT, err := strconv.Atoi(SMTP_PORT_STRING)
	if err != nil {
		log.Fatal("loading smtp port failed")
	}
	SMTP_USERNAME := os.Getenv("SMTP_USERNAME")
	SMTP_PASSWORD := os.Getenv("SMTP_PASSWORD")
	SMTP_SENDER := os.Getenv("SMTP_SENDER")

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment(development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", MOVIE_DB_DSN, "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", SMTP_HOST, "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port",  SMTP_PORT, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", SMTP_USERNAME, "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", SMTP_PASSWORD, "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", SMTP_SENDER, "SMTP sender")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	defer db.Close()

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	err = app.serve()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
