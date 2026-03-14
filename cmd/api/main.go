package main

import (
	"log/slog"
    "strconv"
	"os"
	"time"
	"context"
	"database/sql"

	_ "github.com/lib/pq"
    "github.com/Joseph-Koop/json-project/internal/data"
)

const appVersion = "1.0.0"

type serverConfig struct {
    port int 
    environment string
    db struct {
        dsn string
    }
    limiter struct {
        rps float64                      // requests per second
        burst int                        // initial requests possible
        enabled bool                     // enable or disable rate limiter
    }

}

type applicationDependencies struct {
    config serverConfig
    logger *slog.Logger
    classModel data.ClassModel
}


func main() {
    var settings serverConfig

    settings.environment = os.Getenv("ENVIRONMENT")
    settings.port, _ = strconv.Atoi(os.Getenv("PORT"))
    settings.db.dsn = os.Getenv("DB_DSN")

    settings.limiter.rps, _ = strconv.ParseFloat(os.Getenv("LIMITER_RPS"), 64)
    settings.limiter.burst, _ = strconv.Atoi(os.Getenv("LIMITER_BURST"))
    settings.limiter.enabled, _ = strconv.ParseBool(os.Getenv("LIMITER_ENABLED"))


	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// the call to openDB() sets up our connection pool
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// release the database resources before exiting
	defer db.Close()

	logger.Info("Database connection pool established.")


	appInstance := &applicationDependencies {
        config: settings,
        logger: logger,
        classModel: data.ClassModel {DB: db},
    }

	err = appInstance.serve()
    if err != nil {
        logger.Error(err.Error())
        os.Exit(1)
    }



}  // end of main()

func openDB(settings serverConfig) (*sql.DB, error) {
    // open a connection pool
    db, err := sql.Open("postgres", settings.db.dsn)
    if err != nil {
        return nil, err
    }
    
    // set a context to ensure DB operations don't take too long
    ctx, cancel := context.WithTimeout(context.Background(),
                                       5 * time.Second)
    defer cancel()
    // let's test if the connection pool was created
    // we trying pinging it with a 5-second timeout
    err = db.PingContext(ctx)
    if err != nil {
        db.Close()
        return nil, err
    }


    // return the connection pool (sql.DB)
    return db, nil

} 