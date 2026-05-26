package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/BingyanStudio/is-hust-online/internal/config"
	"github.com/BingyanStudio/is-hust-online/internal/controller/param"
	"github.com/BingyanStudio/is-hust-online/internal/db"
	mymw "github.com/BingyanStudio/is-hust-online/internal/middleware"
	"github.com/BingyanStudio/is-hust-online/internal/service"
	"github.com/BingyanStudio/is-hust-online/internal/views"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"google.golang.org/grpc"
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

	// Context for coordinating shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化任务分发器和调度器
	dispatcher := service.NewTaskDispatcher()
	scheduler := service.NewScheduler(ctx, dispatcher)
	scheduler.Start()

	// 启动 gRPC 服务器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(service.TokenAuthInterceptor()),
		grpc.StreamInterceptor(service.StreamTokenAuthInterceptor()),
	)
	myproto.RegisterClientManagerServer(grpcServer, service.NewClientManagerService(dispatcher))
	myproto.RegisterCheckServiceServer(grpcServer, service.NewCheckServiceService(dispatcher))

	go func() {
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(config.C.GRPCPort))
		if err != nil {
			slog.Error("gRPC 监听失败", "error", err)
			panic(err)
		}
		slog.Warn("gRPC 服务器启动", "port", config.C.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC 服务失败", "error", err)
		}
	}()

	// 启动 HTTP 服务器
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: mymw.CustomHTTPErrorHandler,
		Validator:        param.GetValidator(),
		IPExtractor: func(r *http.Request) string {
			if xff := r.Header.Get(echo.HeaderXForwardedFor); xff != "" {
				// X-Forwarded-For may contain "client, proxy1, proxy2"
				// Take the first IP (the original client)
				for i, c := range xff {
					if c == ',' {
						return xff[:i]
					}
				}
				return xff
			}
			if xri := r.Header.Get(echo.HeaderXRealIP); xri != "" {
				return xri
			}
			return r.RemoteAddr
		},
		Logger: logger,
	})

	// TODO: 改为前端地址
	e.Use(middleware.CORS("https://your-project.bingyan.net"))
	e.Use(middleware.RequestID())

	e.Use(mymw.Logger())

	e.Use(middleware.Recover())

	e.Use(middleware.Gzip())

	views.InitViews(e)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutting down...")

		// Stop scheduler
		scheduler.Stop()
		cancel()

		// Stop gRPC server
		grpcServer.GracefulStop()

		// Close database connections
		if err := db.CloseMongoDB(); err != nil {
			slog.Error("MongoDB close error", "error", err)
		}
	}()

	// 启动服务器
	err = e.Start(":" + strconv.Itoa(config.C.Port))
	if err != nil && err != http.ErrServerClosed {
		slog.Error("服务器启动失败", "error", err)
		panic(err)
	}
}
