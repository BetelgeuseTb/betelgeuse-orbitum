package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/config"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/repository/postgres"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/auth"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/cache"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/oauthflow"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/security"
	httptransport "github.com/BetelgeuseTb/betelgeuse-orbitum/internal/transport/http"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.FromEnv()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	stdDB, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer stdDB.Close()

	rc := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	redisCache := cache.NewRedisCache(rc)

	kp, err := security.NewStaticRSAKeyProviderFromPEM(cfg.JWTKeyPath, cfg.JWTKID)
	if err != nil {
		log.Fatal(err)
	}
	jwtm := security.NewRS256JWTManager(kp)
	hasher := security.NewArgon2idHasher()

	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepoPG(pool)
	accessTokenRepo := postgres.NewAccessTokenRepository(pool)
	revokedRepo := postgres.NewRevokedTokenRepoPG(pool)
	userRoleRepo := postgres.NewUserRoleRepoPG(stdDB)
	authCodeRepo := postgres.NewAuthorizationCodeRepoPG(pool)
	oauthClientRepo := postgres.NewOAuthClientRepoPG(pool)

	authSvc := auth.NewService(userRepo, sessionRepo, accessTokenRepo, revokedRepo, userRoleRepo, jwtm, hasher, redisCache, cfg.Issuer)

	store := oauthflow.NewStore(redisCache)

	handlers := &httptransport.Handlers{
		AuthSvc:         authSvc,
		UserRepo:        userRepo,
		SessionRepo:     sessionRepo,
		OAuthClientRepo: oauthClientRepo,
		AuthCodeRepo:    authCodeRepo,
		AccessTokenRepo: accessTokenRepo,
		RevokedRepo:     revokedRepo,
		RolesRepo:       userRoleRepo,
		OauthStore:      store,
		JWTManager:      jwtm,
		KeyProvider:     kp,
		Hasher:          hasher,
		Issuer:          cfg.Issuer,
	}

	adapter := httptransport.NewAdapter(handlers)
	e := httptransport.NewServer(adapter)

	go func() {
		// We reuse GRPCAddr as the HTTP listen address for consistency with existing config
		if err := e.Start(cfg.GRPCAddr); err != nil {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	_ = e.Shutdown(ctx)
}
