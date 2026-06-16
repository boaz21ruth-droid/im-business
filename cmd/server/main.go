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
	"github.com/web1/im-business/internal/p2p"
	"github.com/web1/im-business/internal/repo"
	"github.com/web1/im-business/internal/service"
	"github.com/web1/im-business/internal/wallet"
	"github.com/web1/im-business/internal/wallet/bridge"
	"github.com/web1/im-business/internal/wallet/intent"
	"github.com/web1/im-business/internal/wallet/provider"
	"github.com/web1/im-business/internal/wallet/quote"
	codepkg "github.com/web1/im-business/pkg/code"
	"github.com/web1/im-business/pkg/cryptoutil"
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
	var smtpCfg *codepkg.SMTPConfig
	if cfg.SMTP != nil {
		smtpCfg = &codepkg.SMTPConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		}
	}
	codeSvc := codepkg.NewService(rdb, smtpCfg)

	accountSvc, err := service.NewAccountService(userRepo, openimCli, jwtSvc, codeSvc)
	if err != nil {
		log.Fatal("account service", zap.Error(err))
	}
	userSvc := service.NewUserService(userRepo)

	totpKey, err := cryptoutil.DecodeKey(cfg.TOTP.EncryptKey)
	if err != nil {
		log.Fatal("totp encrypt_key invalid", zap.Error(err))
	}
	totpSvc := service.NewTotpService(userRepo, rdb, totpKey)

	// Wallet components
	walletRepo := repo.NewWalletRepo(gdb)
	walletCache := wallet.NewCache(rdb)
	tronProv := provider.NewTronGridProvider()
	multiProviders := wallet.BuildProviders(cfg.Wallet, log)
	streamProviders := wallet.BuildStreamProviders(cfg.Wallet, log)
	wsProviders := wallet.BuildWSProviders(cfg.Wallet, log)
	tronPoller := wallet.NewTronPoller(walletRepo, walletCache, openimCli, tronProv, log)
	var localPollers []*wallet.LocalRpcPoller
	for _, lrpc := range cfg.Wallet.LocalRPCs {
		if lrpc.ChainKey == "" || lrpc.RPCURL == "" {
			continue
		}
		lp := wallet.NewLocalRpcPoller(lrpc.ChainKey, lrpc.RPCURL, walletRepo, log)
		localPollers = append(localPollers, lp)
		log.Info("local rpc poller configured", zap.String("chain", lrpc.ChainKey), zap.String("rpc", lrpc.RPCURL))
	}
	walletSvc := wallet.NewWalletService(walletRepo, walletCache, openimCli, streamProviders, wsProviders, multiProviders, tronPoller, localPollers, log)
	swapAggregator := quote.BuildAggregator(cfg.Swap, log)
	bridgeProvider := bridge.BuildProvider(cfg.Bridge)
	intentProvider := intent.BuildProvider(cfg.Intent)

	streamProvidersMap := make(map[string]provider.StreamProvider, len(streamProviders))
	for _, sp := range streamProviders {
		streamProvidersMap[sp.Name()] = sp
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go tronPoller.Start(ctx)
	go walletSvc.StartWSProviders(ctx)
	for _, lp := range localPollers {
		lp.SetOnTransfer(walletSvc.ProcessWebhookTransfer)
	}
	go walletSvc.StartLocalRpcPollers(ctx)

	// P2P escrow
	p2pRepo := repo.NewP2PRepo(gdb)
	if cfg.P2P.ContractAddr != "" {
		p2pListener := p2p.NewListener(cfg.P2P.ContractAddr, p2pRepo, log)
		go p2pListener.Run(ctx)
		log.Info("p2p listener started", zap.String("contract", cfg.P2P.ContractAddr))
	}

	r := handler.NewRouter(log, jwtSvc,
		handler.NewAccountHandler(accountSvc),
		handler.NewUserHandler(userSvc),
		handler.NewWalletHandler(walletSvc, cfg.Swap, swapAggregator, bridgeProvider, intentProvider),
		handler.NewWebhookHandler(walletSvc, streamProvidersMap),
		handler.NewTotpHandler(totpSvc, userSvc),
		handler.NewP2PHandler(p2pRepo, cfg.P2P.ContractAddr, cfg.P2P.ArbitratorKey, log),
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
