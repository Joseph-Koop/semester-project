package main

import (
    "expvar"
    "flag"
	"log/slog"
	"os"
    "runtime"
    "strings"
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
    smtp struct {

    }
    cors struct {
        trustedOrigins []string
    }
}

type applicationDependencies struct {
    config serverConfig
    logger *slog.Logger
    classModel data.ClassModel
    gymModel data.GymModel
    trainerModel data.TrainerModel
    memberModel data.MemberModel
    studioModel data.StudioModel
    sessionTimeModel data.SessionTimeModel
    sessionModel data.SessionModel
    registrationModel data.RegistrationModel
    attendanceModel data.AttendanceModel
    userModel data.UserModel
}


func main() {
    var settings serverConfig

    flag.IntVar(&settings.port, "port", 4000, "Server port")
    flag.StringVar(&settings.environment, "env", "development", "Environment(development|staging|production)")
    // read in the dsn
    flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://gym:gym@localhost/gym?sslmode=disable", "PostgreSQL DSN")

    flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")
    flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")
    flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

    flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)",
        func(val string) error {
            settings.cors.trustedOrigins = strings.Fields(val)
            return nil
        })
    flag.Parse()


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

    // We use the NewString() to provide the key and Set() to specify its value
    expvar.NewString("version").Set(appVersion)

    // the number of active goroutines
    expvar.Publish("goroutines", expvar.Func(func() any {
        return runtime.NumGoroutine()
    }))

    // the database connection pool metrics
    expvar.Publish("database", expvar.Func(func() any {
        return db.Stats()
    }))

    // the current Unix timestamp
    expvar.Publish("timestamp", expvar.Func(func() any {
        return time.Now().Unix()
    }))


	appInstance := &applicationDependencies {
        config: settings,
        logger: logger,
        classModel: data.ClassModel {DB: db},
        gymModel: data.GymModel {DB: db},
        trainerModel: data.TrainerModel {DB: db},
        memberModel: data.MemberModel {DB: db},
        studioModel: data.StudioModel {DB: db},
        sessionTimeModel: data.SessionTimeModel {DB: db},
        sessionModel: data.SessionModel {DB: db},
        registrationModel: data.RegistrationModel {DB: db},
        attendanceModel: data.AttendanceModel {DB: db},
        userModel: data.UserModel {DB: db},
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