package service

import (
	"context"
	"log/slog"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TokenAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod != myproto.ClientManager_Register_FullMethodName {
			return handler(ctx, req)
		}

		token := extractToken(ctx)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		_, err := dao.FindClientByToken(ctx, token)
		if err != nil {
			slog.Warn("auth: invalid token", "method", info.FullMethod)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(ctx, req)
	}
}
