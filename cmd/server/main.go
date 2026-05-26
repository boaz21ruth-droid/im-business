package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/web1/im-business/internal/config"
	"github.com/web1/im-business/internal/db"
	"github.com/web1/im-business/internal/handler"
	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/internal/wallet/provider"
	codepkg "github.com/web1/im-business/pkg/code"
	jwtpkg "github.com/web1/im-business/pkg/jwt"
)

func main() {
	cfgFile := flag.String("config", "config/config.yaml", "config file")
	flag.Parse()

	log, _ := zap.NewProduction()
	defer log.Sync()

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatal("config", zap.Error(err))
	}

	gin.SetMode(cfg.Server.Mode)

	gdb, err := db.NewPostgres(cfg.Postgres)
	if err != nil {
		log.Fatal("postgres", zap.Error(err))
	}

	rdb := db.NewRedis(cfg.Redis)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("redis ping", zap.Error(err))
	}

	userRepo := repo.NewUserRepo(gdb)
	openimCli := service.NewOpenIMClient(cfg.OpenIM)
	jwtSvc := jwtpkg.NewService(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpireHours)*time.Hour)
	codeSvc := codepkg.NewService(rdb)

	accountSvc, err := service.NewAccountService(userRepo, openimCli, jwtSvc, codeSvc)
	if err != nil {
		log.Fatal("account service", zap.Error(err))
	}
	userSvc := service.NewUserService(userRepo)

	// Wallet components
	walletRepo := repo.NewWalletRepo(gdb)
	walletCache := wallet.NewCache(rdb)
	tronProv := provider.NewTronGridProvider()
	multiProviders := wallet.BuildProviders(cfg.Wallet, log)
	streamProviders := wallet.BuildStreamProviders(cfg.Wallet, log)
	wsProviders := wallet.BuildWSProviders(cfg.Wallet, log)
	tronPoller := wallet.NewTronPoller(walletRepo, walletCache, openimCli, tronProv, log)
	walletSvc := wallet.NewWalletService(walletRepo, walletCache, openimCli, streamProviders, wsProviders, multiProviders, tronPoller, log)

	streamProvidersMap := make(map[string]provider.StreamProvider, len(streamProviders))
	for _, sp := range streamProviders {
		streamProvidersMap[sp.Name()] = sp
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go tronPoller.Start(ctx)
	go walletSvc.StartWSProviders(ctx)

	r := handler.NewRouter(log, jwtSvc,
		handler.NewAccountHandler(accountSvc),
		handler.NewUserHandler(userSvc),
		handler.NewWalletHandler(walletSvc),
		handler.NewWebhookHandler(walletSvc, streamProvidersMap),
	)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("im-business starting", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Info("server stopped")
}
