// Command lychee 启动荔枝小秘书后端服务。
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qingwa-ink/lychee/internal/config"
	"github.com/qingwa-ink/lychee/internal/controller"
	"github.com/qingwa-ink/lychee/internal/middleware"
	"github.com/qingwa-ink/lychee/internal/pkg/i18n"
	"github.com/qingwa-ink/lychee/internal/pkg/jwt"
	"github.com/qingwa-ink/lychee/internal/pkg/mail"
	"github.com/qingwa-ink/lychee/internal/repository"
	"github.com/qingwa-ink/lychee/internal/router"
	"github.com/qingwa-ink/lychee/internal/service"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	// 基础组件
	jwtMgr, err := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	if err != nil {
		log.Fatalf("init jwt: %v", err)
	}
	mailer := mail.NewMailer(mail.Config{
		Host: cfg.Mail.Host, Port: cfg.Mail.Port, User: cfg.Mail.User, Pass: cfg.Mail.Pass,
	})
	i18nStore := i18n.New(cfg.I18N.Default)

	// 仓储
	userRepo := repository.NewUserRepository(db)
	codeRepo := repository.NewVerificationRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)

	// 服务 / 控制器 / 中间件
	authSvc := service.NewAuthService(userRepo, codeRepo, refreshRepo, jwtMgr, mailer, 10*time.Minute)
	authCtrl := controller.NewAuthController(authSvc)
	localeCtrl := controller.NewLocaleController(i18nStore, authSvc)

	phraseRepo := repository.NewPhraseRepository(db)
	phraseSvc := service.NewPhraseService(phraseRepo)
	phraseCtrl := controller.NewPhraseController(phraseSvc)

	jwtMW := middleware.JWT(jwtMgr, userRepo)
	jwtOptionalMW := middleware.JWTOptional(jwtMgr, userRepo)
	i18nMW := middleware.I18N(i18nStore)

	r := router.New(cfg, &router.Deps{
		I18NMiddleware:        i18nMW,
		JWTMiddleware:         jwtMW,
		JWTOptionalMiddleware: jwtOptionalMW,
		AuthController:        authCtrl,
		LocaleController:      localeCtrl,
		PhraseController:      phraseCtrl,
	})

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("lychee server listening on %s (env=%s)", addr, cfg.App.Env)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server run: %v", err)
	}
}
