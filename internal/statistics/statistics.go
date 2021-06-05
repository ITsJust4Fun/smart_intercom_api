package statistics

import (
	"context"
	"github.com/pkg/errors"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/auth"
	"smart_intercom_api/internal/report"
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
