package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"
	"fmt"
	"go.keploy.io/server/v2/graph/model"
	"time"

	"go.keploy.io/server/v2/pkg/models"
	"go.keploy.io/server/v2/pkg/service/test"
	"go.keploy.io/server/v2/utils"
	"go.uber.org/zap"
)

// RunTestSet is the resolver for the runTestSet field.
func (r *mutationResolver) RunTestSet(ctx context.Context, testSet string) (*model.RunTestSetResponse, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(Emoji + "failed to get Resolver")
		return nil, err
	}

	tester := r.Resolver.Tester

	if tester == nil {
		r.Logger.Error("failed to get tester from resolver")
		return nil, fmt.Errorf(Emoji+"failed to run testSet:%v", testSet)
	}

	testRunChan := make(chan string, 1)
	pid := r.Resolver.AppPid
	serveTest := r.Resolver.ServeTest
	testCasePath := r.Resolver.Path
	testReportPath := r.Resolver.TestReportPath
	delay := r.Resolver.Delay

	testReportFS := r.Resolver.TestReportFS
	if tester == nil {
		r.Logger.Error("failed to get testReportFS from resolver")
		return nil, fmt.Errorf(Emoji+"failed to run testSet:%v", testSet)
	}

	ys := r.Resolver.Storage
	if ys == nil {
		r.Logger.Error("failed to get ys from resolver")
		return nil, fmt.Errorf(Emoji+"failed to run testSet:%v", testSet)
	}

	loadedHooks := r.LoadedHooks
	if loadedHooks == nil {
		r.Logger.Error("failed to get loadedHooks from resolver")
		return nil, fmt.Errorf(Emoji+"failed to run testSet:%v", testSet)
	}

	resultForTele := make([]int, 2)
	ctx = context.WithValue(ctx, "resultForTele", &resultForTele)
	initialisedValues := test.TestEnvironmentSetup{
		Ctx:            ctx,
		LoadedHooks:    loadedHooks,
		TestReportFS:   testReportFS,
		Storage:        ys,
		IgnoreOrdering: false,
	}
	go func() {
		defer utils.HandlePanic()
		r.Logger.Debug("starting testrun...", zap.Any("testSet", testSet))
		tester.RunTestSet(testSet, testCasePath, testReportPath, "", "", "", delay, 30*time.Second, pid, testRunChan, r.ApiTimeout, nil, nil, serveTest, initialisedValues)
	}()

	testRunID := <-testRunChan
	r.Logger.Debug("", zap.Any("testRunID", testRunID))

	return &model.RunTestSetResponse{Success: true, TestRunID: testRunID}, nil
}

// StopProxy is the resolver for the stopProxy field.
func (r *mutationResolver) StopProxy(ctx context.Context) (bool, error) {
	// if r.Resolver == nil {
	// 	err := fmt.Errorf(Emoji + "failed to get Resolver")
	// 	return false, err
	// }

	// proxy := r.Resolver.Proxy

	// if proxy == nil {
	// 	r.Logger.Error("failed to get proxy from resolver")
	// 	return false, fmt.Errorf(Emoji + "failed to stop the proxy")
	// }

	// proxy.Stop()
	// r.Logger.Debug("proxy stopped")
	// return true, nil
	panic("not implemented")
}

// TestSets is the resolver for the testSets field.
func (r *queryResolver) TestSets(ctx context.Context) ([]string, error) {
	if r.Resolver == nil {
		err := fmt.Errorf(Emoji + "failed to get Resolver")
		return nil, err
	}
	testPath := r.Resolver.Path

	testSets, err := r.Resolver.Storage.ReadTestSessionIndices()
	if err != nil {
		r.Resolver.Logger.Error("failed to fetch test sets", zap.Any("testPath", testPath), zap.Error(err))
		return nil, err
	}

	// Print debug log for retrieved qualified test sets
	if len(testSets) > 0 {
		r.Resolver.Logger.Debug(fmt.Sprintf("Retrieved test sets: %v", testSets), zap.Any("testPath", testPath))
	} else {
		r.Resolver.Logger.Debug("No test sets found", zap.Any("testPath", testPath))
	}

	return testSets, nil
}

// TestSetStatus is the resolver for the testSetStatus field.
func (r *queryResolver) TestSetStatus(ctx context.Context, testRunID string) (*model.TestSetStatus, error) {
	//Initiate the telemetry.
	// var tele = telemetry.NewTelemetry(true, false, store, r.Logger, "", nil)

	// var tele = telemetry.NewTelemetry(true, false, store, r.Logger, "", nil)

	if r.Resolver == nil {
		err := fmt.Errorf(Emoji + "failed to get Resolver")
		return nil, err
	}
	testReportFs := r.Resolver.TestReportFS

	if testReportFs == nil {
		r.Logger.Error("failed to get testReportFS from resolver")
		return nil, fmt.Errorf(Emoji+"failed to get the status for testRunID:%v", testRunID)
	}
	testReport, err := testReportFs.Read(ctx, r.Resolver.TestReportPath, testRunID)
	if err != nil {
		r.Logger.Error("failed to fetch testReport", zap.Any("testRunID", testRunID), zap.Error(err))
		return nil, err
	}
	readTestReport, ok := testReport.(*models.TestReport)
	if !ok {
		r.Logger.Error("failed to read testReport from resolver")
		return nil, fmt.Errorf(Emoji+"failed to read the test report for testRunID:%v", testRunID)
	}
	if readTestReport.Status == "PASSED" || readTestReport.Status == "FAILED" {
		tele.Testrun(readTestReport.Success, readTestReport.Failure)
	}

	r.Logger.Debug("", zap.Any("testRunID", testRunID), zap.Any("testSetStatus", readTestReport.Status))
	return &model.TestSetStatus{Status: readTestReport.Status}, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
