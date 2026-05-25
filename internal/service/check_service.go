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

	check := &model.Check{
		SiteID:   result.Id,
		ClientID: req.ClientId,
		Type:     result.CheckType,
		Status:   result.ErrorType,
		Result:   buildResultString(result),
		Delay:    int64(responseTimeMs),
	}

	if err := dao.InsertCheck(ctx, check); err != nil {
		slog.Error("failed to insert check", "error", err)
		return nil, status.Error(codes.Internal, "failed to save check result")
	}

	// Update report aggregation
	if err := updateReports(ctx, result.Id, result.Success, float64(responseTimeMs)); err != nil {
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

func updateReports(ctx context.Context, siteID string, success bool, delay float64) error {
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
			SiteID:    siteID,
			Timeframe: rt.timeframe,
			Type:      rt.reportType,
			Successes: successCount,
			Uptime:    0,
			AvgDelay:  delay,
		}

		if err := dao.UpsertReport(ctx, report); err != nil {
			return err
		}

		// Recalculate uptime percentage after incrementing counters
		RecalculateReportUptime(ctx, siteID, rt.timeframe, rt.reportType)
	}
	return nil
}

// RecalculateReportUptime recalculates the uptime percentage for a report
// based on its current checks and successes counts.
// This is called after UpsertReport to fix the uptime value.
func RecalculateReportUptime(ctx context.Context, siteID, timeframe string, reportType int) {
	rpt, err := dao.FindReport(ctx, siteID, timeframe, reportType)
	if err != nil || rpt == nil || rpt.Checks == 0 {
		return
	}
	uptime := (float64(rpt.Successes) / float64(rpt.Checks)) * 100
	uptime = math.Round(uptime*100) / 100
	dao.SetReportUptime(ctx, siteID, timeframe, reportType, uptime)
}
