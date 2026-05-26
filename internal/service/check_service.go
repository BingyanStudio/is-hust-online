package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/dao"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	myproto "github.com/BingyanStudio/is-hust-online/pkg/proto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CheckServiceService struct {
	myproto.UnimplementedCheckServiceServer
	dispatcher *TaskDispatcher
}

func NewCheckServiceService(dispatcher *TaskDispatcher) *CheckServiceService {
	return &CheckServiceService{dispatcher: dispatcher}
}

func (s *CheckServiceService) WatchTasks(req *myproto.WatchTasksRequest, stream myproto.CheckService_WatchTasksServer) error {
	clientID := req.ClientId
	taskCh, exists := s.dispatcher.GetChannel(clientID)
	if !exists {
		return status.Error(codes.NotFound, "client not registered")
	}

	slog.Info("client watching tasks", "client_id", clientID)

	defer func() {
		s.dispatcher.UnregisterClient(clientID)
		slog.Info("client stopped watching tasks", "client_id", clientID)
	}()

	for {
		select {
		case task, ok := <-taskCh:
			if !ok {
				return nil
			}
			if err := stream.Send(task); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func (s *CheckServiceService) ReportResult(ctx context.Context, req *myproto.CheckResultRequest) (*myproto.CheckResultResponse, error) {
	result := req.Result
	if result == nil {
		return nil, status.Error(codes.InvalidArgument, "missing result")
	}

	var responseTimeMs int32
	if result.ResponseTimeMs != nil {
		responseTimeMs = result.GetResponseTimeMs()
	}

	siteID, err := bson.ObjectIDFromHex(result.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid site id")
	}
	clientID, err := bson.ObjectIDFromHex(req.ClientId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid client id")
	}

	checkConfigID, _ := bson.ObjectIDFromHex(result.CheckConfigId)

	check := &model.Check{
		SiteID:        siteID,
		ClientID:      clientID,
		CheckConfigID: checkConfigID,
		Type:          result.CheckType,
		Status:        result.ErrorType,
		Result:        buildResultString(result),
		Delay:         int64(responseTimeMs),
	}

	if err := dao.InsertCheck(ctx, check); err != nil {
		slog.Error("failed to insert check", "error", err)
		return nil, status.Error(codes.Internal, "failed to save check result")
	}

	if err := updateReports(ctx, siteID, checkConfigID, result.Success, float64(responseTimeMs)); err != nil {
		slog.Error("failed to update report", "error", err)
	}

	return &myproto.CheckResultResponse{Success: true}, nil
}

func buildResultString(resp *myproto.CheckResponse) string {
	if resp.Success {
		return "success"
	}
	return fmt.Sprintf("error: %s", resp.ErrorType.String())
}

func updateReports(ctx context.Context, siteID, checkConfigID bson.ObjectID, success bool, delay float64) error {
	now := time.Now()
	successCount := int64(0)
	if success {
		successCount = 1
	}

	reportTypes := []struct {
		reportType int
		timeframe  string
	}{
		{model.REPORT_TYPE_HOURLY, now.Format("2006-01-02 15:00:00")},
		{model.REPORT_TYPE_DAILY, now.Format("2006-01-02")},
		{model.REPORT_TYPE_MONTHLY, now.Format("2006-01")},
	}

	for _, rt := range reportTypes {
		report := &model.Report{
			SiteID:        siteID,
			CheckConfigID: checkConfigID,
			Timeframe:     rt.timeframe,
			Type:          rt.reportType,
			Successes:     successCount,
			Uptime:        0,
			AvgDelay:      delay,
		}

		if err := dao.UpsertReport(ctx, report); err != nil {
			return err
		}

		RecalculateReportUptime(ctx, siteID, checkConfigID, rt.timeframe, rt.reportType)
	}
	return nil
}

func RecalculateReportUptime(ctx context.Context, siteID, checkConfigID bson.ObjectID, timeframe string, reportType int) {
	rpt, err := dao.FindReport(ctx, siteID, checkConfigID, timeframe, reportType)
	if err != nil || rpt == nil || rpt.Checks == 0 {
		return
	}
	uptime := (float64(rpt.Successes) / float64(rpt.Checks)) * 100
	uptime = math.Round(uptime*100) / 100
	dao.SetReportUptime(ctx, siteID, checkConfigID, timeframe, reportType, uptime)
}
