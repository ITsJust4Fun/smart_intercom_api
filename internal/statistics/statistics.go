package statistics

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/internal/report"
	"smart_intercom_api/pkg/config"
	pb "smart_intercom_api/proto"
)

func ReportStatisticsQuery(ctx context.Context) (*model.ReportStatistics, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	allReports, err := report.GetAll()

	if err != nil {
		return nil, err
	}

	reportStatistics := model.ReportStatistics {
		Normal: 0,
		Warnings: 0,
		Errors: 0,
	}

	for _, reportData := range allReports {
		switch reportData.Level {
		case 0:
			reportStatistics.Normal++
		case 1:
			reportStatistics.Warnings++
		case 2:
			reportStatistics.Errors++
		}
	}

	return &reportStatistics, nil
}

func HardwareStatisticsQuery(ctx context.Context) (*model.HardwareStatistics, error) {
	if !auth.GetLoginState(ctx) {
		return nil, errors.New("access denied")
	}

	connect, err := grpc.Dial(config.GetConfig().DiagnosticsProto, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		return nil, err
	}

	defer func(connect *grpc.ClientConn) {
		err := connect.Close()

		if err != nil {
			return
		}
	}(connect)

	client := pb.NewDiagnosticsClient(connect)
	ctx, cancel := context.WithTimeout(context.Background(), config.GetConfig().DatabaseTimeout)
	defer cancel()

	result, err := client.GetDiagnostic(ctx, &pb.Empty{})

	if err != nil {
		return nil, err
	}

	hardwareStatistics := model.HardwareStatistics{
		CPUUsage: result.Cpu,
		FreeRAM: result.FreeRAM,
		UsedRAM: result.UsedRAM,
		TotalRAM: result.TotalRAM,
		FreeHdd: result.FreeHDD,
		UsedHdd: result.UsedHDD,
		TotalHdd: result.TotalHDD,
	}

	return &hardwareStatistics, nil
}
