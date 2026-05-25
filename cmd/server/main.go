package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"

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

	err = db.InitRedisClient(config.C.Redis)
	if err != nil {
		slog.Error("Redis 客户端初始化失败", "error", err)
		panic(err)
	}
	slog.Warn("Redis 连接成功")

	// 初始化任务分发器和调度器
	dispatcher := service.NewTaskDispatcher()
	scheduler := service.NewScheduler(dispatcher)
	scheduler.Start(context.Background())

	// 启动 gRPC 服务器
	go func() {
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(config.C.GRPCPort))
		if err != nil {
			slog.Error("gRPC 监听失败", "error", err)
			panic(err)
		}

		s := grpc.NewServer(
			grpc.UnaryInterceptor(service.TokenAuthInterceptor()),
		)

		myproto.RegisterClientManagerServer(s, service.NewClientManagerService(dispatcher))
		myproto.RegisterCheckServiceServer(s, service.NewCheckServiceService(dispatcher))

		slog.Warn("gRPC 服务器启动", "port", config.C.GRPCPort)
		if err := s.Serve(lis); err != nil {
			slog.Error("gRPC 服务失败", "error", err)
		}
	}()

	// 启动 HTTP 服务器
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

	// TODO: 改为前端地址
	e.Use(middleware.CORS("https://your-project.bingyan.net"))
	e.Use(middleware.RequestID())

	e.Use(mymw.Logger())

	e.Use(middleware.Recover())

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
