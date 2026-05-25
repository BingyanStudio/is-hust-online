package main

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/BingyanStudio/is-hust-online/internal/config"
	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/db"
	mymw "github.com/BingyanStudio/is-hust-online/internal/middleware"
	"github.com/BingyanStudio/is-hust-online/internal/views"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// 初始化配置
	err := config.LoadConfig()
	if err != nil {
		slog.Error("配置初始化失败", "error", err)
		panic(err)
	}

	// 初始化日志
	level := slog.LevelWarn
	if config.C.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))
	slog.SetDefault(logger)

	// 初始化数据库
	err = db.InitMongoDB(config.C.Mongo)
	if err != nil {
		slog.Error("数据库初始化失败", "error", err)
		panic(err)
	}
	slog.Warn("MongoDB 连接成功")

	err = db.InitRedisClient(config.C.Redis)
	if err != nil {
		slog.Error("Redis 客户端初始化失败", "error", err)
		panic(err)
	}
	slog.Warn("Redis 连接成功")

	// 启动数据维护和修补任务

	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: mymw.CustomHTTPErrorHandler,
		Validator:        param.GetValidator(),
		IPExtractor: func(r *http.Request) string {
			xri := r.Header.Get(echo.HeaderXRealIP)
			if xri != "" {
				return xri
			}
			xff := r.Header.Get(echo.HeaderXForwardedFor)
			if xff != "" {
				return xff
			}
			return r.RemoteAddr
		},
		Logger: logger,
	})

	sampleRate := 0.3
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.C.SentryDsn,
		EnableTracing:    true,
		TracesSampleRate: sampleRate,
		SendDefaultPII:   true,
		EnableLogs:       true,
	}); err != nil {
		slog.Error("Sentry initialization failed", "error", err)
	}

	// TODO: 改为前端地址
	e.Use(middleware.CORS("https://your-project.bingyan.net"))
	e.Use(middleware.RequestID())

	e.Use(mymw.Logger())

	e.Use(middleware.Recover())
	e.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))
	e.Use(mymw.SentryUserMiddleware())

	e.Use(middleware.Gzip())

	// defer db.CloseMongoDB()
	// defer db.CloseRedisClient()

	views.InitViews(e)

	// 启动服务器
	err = e.Start(":" + strconv.Itoa(config.C.Port))
	if err != nil {
		slog.Error("服务器启动失败", "error", err)
		panic(err)
	}
}
