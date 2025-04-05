package main

import (
	"social/internal/auth"
	"social/internal/db"
	"social/internal/env"
	"social/internal/mailer"
	"social/internal/store"
	"social/internal/store/cache"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			GoSocial API
//	@description	This is a server for Go Devs
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

//	@securityDefinitions.apiKey	ApiKeyAuth
//	@in							header
//	@name						Authorization

func main() {

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "http://localhost:8080"),
		dbconn: dbconfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost:5433/social_network?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 25),
			maxIdletime:  env.GetString("DB_MAX_OPEN_CONNS_LIFETIME", "15m"),
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PASSWORD", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env : env.GetString("ENV", "development"),
		mail : mailconfig{
			exp : time.Hour * 24 * 3,
			fromEmail: env.GetString("SENDGRID_FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
		},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		auth: authconfig{
			basic: basicconfig{
				user: env.GetString("BASIC_AUTH_USER", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "password"),
			},
			token : tokenconfig{
				secret: env.GetString("JWT_SECRET", ""),
				aud: env.GetString("JWT_AUD", "gosocial"),
				iss: env.GetString("JWT_ISS", "gosocial"),
				exp: time.Hour * 24 * 7,
			},
		},
	}

	//Logger 
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		cfg.dbconn.addr,
		cfg.dbconn.maxOpenConns,
		cfg.dbconn.maxIdleConns,
		cfg.dbconn.maxIdletime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("Database connection established")

	var rdb *redis.Client

	if cfg.redisCfg.enabled {
		rdb = redis.NewClient(&redis.Options{
			Addr: cfg.redisCfg.addr,
			Password: cfg.redisCfg.pw,
			DB: cfg.redisCfg.db,
		})

		logger.Info("Redis connection established")
	}

	store := store.NewStorage(db)

	cacheStorage := cache.NewRedisStore(rdb)

	mailer := mailer.NewSendGridMailer(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	jwtAuth := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.aud)

	app := &application{
		config: cfg,
		store:  store,
		cacheStorage : cacheStorage,
		logger: logger,	
		mailer: mailer,	
		authenticator: jwtAuth,

	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
