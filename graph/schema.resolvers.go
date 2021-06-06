package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"smart_intercom_api/graph/generated"
	"smart_intercom_api/graph/model"
	"smart_intercom_api/internal/login"
	"smart_intercom_api/internal/report"
	"smart_intercom_api/internal/statistics"
	"smart_intercom_api/internal/videos"
)

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	return login.LoginMutation(ctx, input)
}

func (r *mutationResolver) ChangePassword(ctx context.Context, input model.NewPassword) (string, error) {
	return login.ChangePasswordMutation(ctx, input)
}

func (r *mutationResolver) CreateVideo(ctx context.Context, input model.NewVideo) (*model.Video, error) {
	return videos.CreateVideoMutation(ctx, input)
}

func (r *mutationResolver) RemoveVideo(ctx context.Context, input model.RemoveVideo) (*model.Video, error) {
	return videos.RemoveVideoMutation(ctx, input)
}

func (r *mutationResolver) CreateReport(ctx context.Context, input model.NewReport) (*model.Report, error) {
	return report.CreateReportMutation(ctx, input)
}

func (r *mutationResolver) ViewReport(ctx context.Context, input model.ViewReport) (*model.Report, error) {
	return report.ViewReportMutation(ctx, input)
}

func (r *mutationResolver) RemoveReport(ctx context.Context, input model.RemoveReport) (*model.Report, error) {
	return report.RemoveReportMutation(ctx, input)
}

func (r *queryResolver) Videos(ctx context.Context) ([]*model.Video, error) {
	return videos.Query(ctx)
}

func (r *queryResolver) Reports(ctx context.Context) ([]*model.Report, error) {
	return report.ReportsQuery(ctx)
}

func (r *queryResolver) UnviewedReportsCount(ctx context.Context) (int, error) {
	return report.UnviewedReportsCount(ctx)
}

func (r *queryResolver) HardwareStatistics(ctx context.Context) (*model.HardwareStatistics, error) {
	return statistics.HardwareStatisticsQuery(ctx)
}

func (r *queryResolver) ReportStatistics(ctx context.Context) (*model.ReportStatistics, error) {
	return statistics.ReportStatisticsQuery(ctx)
}

func (r *queryResolver) RefreshToken(ctx context.Context) (string, error) {
	return login.RefreshTokenQuery(ctx)
}

func (r *queryResolver) Logout(ctx context.Context) (string, error) {
	return login.LogoutQuery(ctx)
}

func (r *subscriptionResolver) VideoUpdated(ctx context.Context) (<-chan *model.Video, error) {
	return videos.VideoUpdatedSubscription(ctx)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
