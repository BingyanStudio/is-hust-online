package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	serverAddr := flag.String("server", "localhost:9090", "gRPC server address")
	token := flag.String("token", "", "authentication token")
	capabilities := flag.String("capabilities", "http,ping,tcp", "comma-separated list of check capabilities (http, ping, tcp)")
	insecureFlag := flag.Bool("insecure", false, "use insecure gRPC connection (no TLS)")
	flag.Parse()

	if *token == "" {
		slog.Error("token is required")
		os.Exit(1)
	}

	var creds credentials.TransportCredentials
	if *insecureFlag {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	}

	conn, err := grpc.NewClient(*serverAddr,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	cmClient := myproto.NewClientManagerClient(conn)
	csClient := myproto.NewCheckServiceClient(conn)

	// Register with token in metadata
	ip, err := retrieveIPv4Address()
	if err != nil {
		slog.Warn("failed to retrieve public IP, using local hostname", "error", err)
		ip = "unknown"
	}
	md := metadata.Pairs("authorization", "Bearer "+*token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	caps := []myproto.CheckType{}
	for _, cap := range splitAndTrim(*capabilities) {
		switch cap {
		case "http":
			caps = append(caps, myproto.CheckType_CHECK_TYPE_HTTP)
		case "ping":
			caps = append(caps, myproto.CheckType_CHECK_TYPE_PING)
		case "tcp":
			caps = append(caps, myproto.CheckType_CHECK_TYPE_TCP)
		}
	}

	regResp, err := cmClient.Register(ctx, &myproto.RegisterRequest{
		ClientInfo: &myproto.ClientInfo{
			Ip:           ip,
			Capabilities: caps,
		},
	})
	if err != nil {
		slog.Error("failed to register", "error", err)
		os.Exit(1)
	}
	if !regResp.Success {
		slog.Error("registration rejected", "message", regResp.Message)
		os.Exit(1)
	}

	clientID := regResp.ClientId
	slog.Info("registered", "client_id", clientID)

	// Heartbeat goroutine
	heartbeatCtx, heartbeatCancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hbCtx := metadata.NewOutgoingContext(heartbeatCtx, md)
				_, err := cmClient.Heartbeat(hbCtx, &myproto.HeartbeatRequest{
					ClientId: clientID,
					Status:   myproto.ClientStatus_CLIENT_STATUS_ONLINE,
				})
				if err != nil {
					slog.Warn("heartbeat failed", "error", err)
				}
			case <-heartbeatCtx.Done():
				return
			}
		}
	}()

	// WatchTasks stream
	go func() {
		for {
			streamCtx := metadata.NewOutgoingContext(context.Background(), md)
			stream, err := csClient.WatchTasks(streamCtx, &myproto.WatchTasksRequest{
				ClientId: clientID,
			})
			if err != nil {
				slog.Error("failed to watch tasks", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for {
				task, err := stream.Recv()
				if err != nil {
					slog.Warn("stream ended", "error", err)
					break
				}
				go performCheck(csClient, clientID, *token, task)
			}
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutting down...")
	heartbeatCancel()

	deregCtx := metadata.NewOutgoingContext(context.Background(), md)
	cmClient.Deregister(deregCtx, &myproto.DeregisterRequest{ClientId: clientID})
	slog.Info("deregistered")
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i, v := range parts {
		parts[i] = strings.TrimSpace(v)
	}
	return parts
}

func retrieveIPv4Address() (string, error) {
	// retrieve using ip.sb
	resp, err := http.Get("https://api-ipv4.ip.sb/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(data))
	return ip, nil
}

func performCheck(csClient myproto.CheckServiceClient, clientID string, token string, task *myproto.CheckTask) {
	start := time.Now()
	check := task.Check
	if check == nil {
		return
	}

	var resp *myproto.CheckResponse

	switch check.CheckType {
	case myproto.CheckType_CHECK_TYPE_HTTP:
		resp = doHTTPCheck(check)
	case myproto.CheckType_CHECK_TYPE_PING:
		resp = doPingCheck(check)
	case myproto.CheckType_CHECK_TYPE_TCP:
		resp = doTCPCheck(check)
	default:
		resp = &myproto.CheckResponse{
			Id:            check.Id,
			Success:       false,
			ErrorType:     myproto.ErrorType_ERROR_TYPE_NO_ERROR,
			CheckType:     check.CheckType,
			CheckConfigId: check.CheckConfigId,
		}
	}

	resp.CheckType = check.CheckType
	resp.CheckConfigId = check.CheckConfigId
	elapsed := int32(time.Since(start).Milliseconds())
	resp.ResponseTimeMs = &elapsed
	resp.Timestamp = timestamppb.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	_, err := csClient.ReportResult(ctx, &myproto.CheckResultRequest{
		ClientId: clientID,
		TaskId:   task.TaskId,
		Result:   resp,
	})
	if err != nil {
		slog.Error("failed to report result", "error", err, "task_id", task.TaskId)
	}
}

func doHTTPCheck(check *myproto.CheckRequest) *myproto.CheckResponse {
	timeout := 10 * time.Second
	if check.TimeoutSeconds != nil {
		timeout = time.Duration(check.GetTimeoutSeconds()) * time.Second
	}

	client := &http.Client{Timeout: timeout}
	method := check.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, check.Url, nil)
	if err != nil {
		return &myproto.CheckResponse{
			Id:        check.Id,
			Success:   false,
			ErrorType: myproto.ErrorType_ERROR_TYPE_HTTP_OTHER,
		}
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36 Edg/149.0.0.0")

	resp, err := client.Do(req)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &myproto.CheckResponse{
				Id:        check.Id,
				Success:   false,
				ErrorType: myproto.ErrorType_ERROR_TYPE_HTTP_TIMEOUT,
			}
		}
		return &myproto.CheckResponse{
			Id:        check.Id,
			Success:   false,
			ErrorType: myproto.ErrorType_ERROR_TYPE_HTTP_UNREACHABLE,
		}
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	extra, err := json.Marshal(map[string]string{"status_code": http.StatusText(resp.StatusCode)})
	if err != nil {
		extra = []byte("{}")
	}

	return &myproto.CheckResponse{
		Id:        check.Id,
		Success:   success,
		ErrorType: myproto.ErrorType_ERROR_TYPE_NO_ERROR,
		Extra:     extra,
	}
}

func doPingCheck(check *myproto.CheckRequest) *myproto.CheckResponse {
	// Simplified: use TCP dial to port 80 as proxy for ping
	host := check.Url
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "80")
	}

	timeout := 5 * time.Second
	if check.TimeoutSeconds != nil {
		timeout = time.Duration(check.GetTimeoutSeconds()) * time.Second
	}

	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &myproto.CheckResponse{
				Id:        check.Id,
				Success:   false,
				ErrorType: myproto.ErrorType_ERROR_TYPE_PING_TIMEOUT,
			}
		}
		return &myproto.CheckResponse{
			Id:        check.Id,
			Success:   false,
			ErrorType: myproto.ErrorType_ERROR_TYPE_PING_UNREACHABLE,
		}
	}
	conn.Close()

	return &myproto.CheckResponse{
		Id:        check.Id,
		Success:   true,
		ErrorType: myproto.ErrorType_ERROR_TYPE_NO_ERROR,
	}
}

func doTCPCheck(check *myproto.CheckRequest) *myproto.CheckResponse {
	host := check.Url
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "80")
	}

	timeout := 5 * time.Second
	if check.TimeoutSeconds != nil {
		timeout = time.Duration(check.GetTimeoutSeconds()) * time.Second
	}

	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &myproto.CheckResponse{
				Id:        check.Id,
				Success:   false,
				ErrorType: myproto.ErrorType_ERROR_TYPE_TCP_TIMEOUT,
			}
		}
		return &myproto.CheckResponse{
			Id:        check.Id,
			Success:   false,
			ErrorType: myproto.ErrorType_ERROR_TYPE_TCP_UNREACHABLE,
		}
	}
	conn.Close()

	return &myproto.CheckResponse{
		Id:        check.Id,
		Success:   true,
		ErrorType: myproto.ErrorType_ERROR_TYPE_NO_ERROR,
	}
}
