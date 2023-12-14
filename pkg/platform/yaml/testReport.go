package yaml

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.keploy.io/server/pkg/models"
	"go.keploy.io/server/pkg/platform"
	"go.keploy.io/server/pkg/proxy/util"
	"go.uber.org/zap"
	yamlLib "gopkg.in/yaml.v3"
)

type TestReport struct {
	tests  map[string][]platform.KindSpecifier
	m      sync.Mutex
	Logger *zap.Logger
}

func NewTestReportFS(logger *zap.Logger) *TestReport {
	return &TestReport{
		tests:  make(map[string][]platform.KindSpecifier), // Correctly initialize the map
		m:      sync.Mutex{},
		Logger: logger,
	}
}

func (fe *TestReport) Lock() {
	fe.M.Lock()
}

func (fe *TestReport) Unlock() {
	fe.M.Unlock()
}

func (fe *TestReport) SetResult(runId string, test platform.KindSpecifier) {
	fe.m.Lock()
	tests := fe.tests[runId]
	tests = append(tests, test)
	fe.Tests[runId] = tests
	fe.M.Unlock()
}

func (fe *TestReport) GetResults(runId string) ([]platform.KindSpecifier, error) {
	testResults, ok := fe.tests[runId]
	if !ok {
		return nil, fmt.Errorf("%s found no test results for test report with id: %s", Emoji, runId)
	}

	return testResults, nil
}

func (fe *TestReport) Read(ctx context.Context, path, name string) (platform.KindSpecifier, error) {

	file, err := os.OpenFile(filepath.Join(path, name+".yaml"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return &models.TestReport{}, err
	}
	defer file.Close()
	decoder := yamlLib.NewDecoder(file)
	var doc models.TestReport
	err = decoder.Decode(&doc)
	if err != nil {
		return &models.TestReport{}, fmt.Errorf(Emoji, "failed to decode the yaml file documents. error: %v", err.Error())
	}
	return &doc, nil
}

func (fe *TestReport) Write(ctx context.Context, path string, doc platform.KindSpecifier) error {
	readDock, ok := doc.(*models.TestReport)
	if !ok {
		return fmt.Errorf(Emoji, "failed to read test report in yaml file.")
	}
	if readDock.Name == "" {
		lastIndex, err := findLastIndex(path, fe.Logger)
		if err != nil {
			return err
		}
		readDock.Name = fmt.Sprintf("report-%v", lastIndex)
	}

	_, err := util.CreateYamlFile(path, readDock.Name, fe.Logger)
	if err != nil {
		return err
	}

	data := []byte{}
	d, err := yamlLib.Marshal(&doc)
	if err != nil {
		return fmt.Errorf(Emoji, "failed to marshal document to yaml. error: %s", err.Error())
	}
	data = append(data, d...)

	err = os.WriteFile(filepath.Join(path, readDock.Name+".yaml"), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf(Emoji, "failed to write test report in yaml file. error: %s", err.Error())
	}
	return nil
}
