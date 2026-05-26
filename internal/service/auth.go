package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TokenAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

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

func StreamTokenAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if info.FullMethod == myproto.CheckService_WatchTasks_FullMethodName {
			token := extractToken(ss.Context())
			if token == "" {
				return status.Error(codes.Unauthenticated, "missing authorization token")
			}

			_, err := dao.FindClientByToken(ss.Context(), token)
			if err != nil {
				slog.Warn("auth: invalid token", "method", info.FullMethod)
				return status.Error(codes.Unauthenticated, "invalid token")
			}
		}

		return handler(srv, ss)
	}
}

func extractToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return ""
	}
	parts := strings.SplitN(authHeaders[0], " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return authHeaders[0]
}
