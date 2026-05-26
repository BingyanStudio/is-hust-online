package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ClientManagerService struct {
	myproto.UnimplementedClientManagerServer
	dispatcher *TaskDispatcher
}

func NewClientManagerService(dispatcher *TaskDispatcher) *ClientManagerService {
	return &ClientManagerService{dispatcher: dispatcher}
}

func (s *ClientManagerService) Register(ctx context.Context, req *myproto.RegisterRequest) (*myproto.RegisterResponse, error) {
	token := extractToken(ctx)
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	client, err := dao.FindClientByToken(ctx, token)
	if err != nil {
		return &myproto.RegisterResponse{
			Success: false,
			Message: "invalid token",
		}, status.Error(codes.Unauthenticated, "invalid token")
	}

	peerIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		peerIP = p.Addr.String()
	}

	cap := 0

	for _, c := range req.ClientInfo.Capabilities {
		cap |= int(c)
	}

	err = dao.UpdateClient(ctx, client.ID, bson.M{
		"status":       0, // CLIENT_STATUS_ONLINE
		"ip":           req.ClientInfo.Ip,
		"last_online":  time.Now().Unix(),
		"capabilities": cap,
	})
	if err != nil {
		slog.Error("failed to update client on register", "error", err, "client_id", client.ID.Hex())
		return nil, status.Error(codes.Internal, "failed to register")
	}

	clientID := client.ID.Hex()
	s.dispatcher.RegisterClient(clientID, client.Capabilities)

	slog.Info("client registered", "client_id", clientID, "name", client.Name, "ip", peerIP)

	return &myproto.RegisterResponse{
		ClientId: clientID,
		Success:  true,
		Message:  "registered successfully",
	}, nil
}

func (s *ClientManagerService) Heartbeat(ctx context.Context, req *myproto.HeartbeatRequest) (*myproto.HeartbeatResponse, error) {
	clientID, err := bson.ObjectIDFromHex(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client_id")
	}

	err = dao.UpdateClient(ctx, clientID, bson.M{
		"status":      int(req.Status),
		"last_online": time.Now().Unix(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update heartbeat")
	}

	return &myproto.HeartbeatResponse{
		Success:    true,
		ServerTime: time.Now().Unix(),
	}, nil
}

func (s *ClientManagerService) Deregister(ctx context.Context, req *myproto.DeregisterRequest) (*myproto.DeregisterResponse, error) {
	clientID, err := bson.ObjectIDFromHex(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client_id")
	}

	err = dao.UpdateClient(ctx, clientID, bson.M{
		"status": 1, // CLIENT_STATUS_OFFLINE
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to deregister")
	}

	s.dispatcher.UnregisterClient(req.ClientId)

	slog.Info("client deregistered", "client_id", req.ClientId)

	return &myproto.DeregisterResponse{Success: true}, nil
}

