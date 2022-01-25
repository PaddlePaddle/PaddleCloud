// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/driver"
)

type Server struct {
	driver.Driver
	Log    logr.Logger
	ctx    context.Context
	cancel context.CancelFunc

	watcher   *fsnotify.Watcher
	svrOpt    *common.ServerOptions
	rootOpt   *common.RootCmdOptions
	doers     map[string]func([]byte) error
	optResMap map[string]string
}

func NewServer(rootOpt *common.RootCmdOptions, svrOpt *common.ServerOptions) (*Server, error) {
	driverName := v1alpha1.DriverName(rootOpt.Driver)
	csiDriver, err := driver.GetDriver(driverName)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(svrOpt.ServerDir, "/") {
		return nil, fmt.Errorf("path must begin with /")
	}

	// configure zap log and create a logger
	zapLog := zap.New(func(o *zap.Options) {
		o.Development = rootOpt.Development
	}, func(o *zap.Options) {
		o.ZapOpts = append(o.ZapOpts, zapOpt.AddCaller())
	},
		func(o *zap.Options) {
			if !rootOpt.Development {
				encCfg := zapOpt.NewProductionEncoderConfig()
				encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
				encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
				o.Encoder = zapcore.NewConsoleEncoder(encCfg)
			}
		})

	// create file system notify watcher and add dir to watch
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// make maps
	optResMap := make(map[string]string)
	doer := make(map[string]func([]byte) error)

	// make Cancel Context
	ctx, cancel := context.WithCancel(context.Background())

	server := &Server{
		ctx:       ctx,
		cancel:    cancel,
		doers:     doer,
		rootOpt:   rootOpt,
		svrOpt:    svrOpt,
		Log:       zapLog,
		watcher:   watcher,
		Driver:    csiDriver,
		optResMap: optResMap,
	}
	return server, nil
}

func (s *Server) Run() error {
	defer s.cancel()
	defer s.watcher.Close()

	// add job doer for watcher's event
	s.doers[common.PathSyncOptions] = s.doSync
	s.doers[common.PathClearOptions] = s.doClear
	s.doers[common.PathRmrOptions] = s.doRmr
	s.doers[common.PathWarmupOptions] = s.doWarmup
	// add options to result map key-value pair
	s.optResMap[common.PathSyncOptions] = common.PathSyncResult
	s.optResMap[common.PathClearOptions] = common.PathClearResult
	s.optResMap[common.PathRmrOptions] = common.PathRmrResult
	s.optResMap[common.PathWarmupOptions] = common.PathWarmupResult

	// add static file server handlers
	if err := s.addStaticHandlers(
		common.PathCacheStatus,
		common.PathSyncResult,
		common.PathClearResult,
		common.PathRmrResult,
		common.PathWarmupResult,

		common.PathSyncOptions,
		common.PathClearOptions,
		common.PathRmrOptions,
		common.PathWarmupOptions,
	); err != nil {
		return err
	}

	// add upload option file handlers
	s.addUploadHandlers(
		common.PathSyncOptions,
		common.PathClearOptions,
		common.PathRmrOptions,
		common.PathWarmupOptions)

	if err := s.addWatchDirs(
		common.PathSyncOptions,
		common.PathClearOptions,
		common.PathRmrOptions,
		common.PathWarmupOptions,
	); err != nil {
		return err
	}

	// (TODO: xiaolao) Add one goroutine to clear options and results file
	go s.watchAndDo()
	go s.writeCacheStatus()

	s.Log.V(1).Info("===== run server ======")
	addr := fmt.Sprintf(":%d", common.RuntimeServicePort)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) uploadHandleFunc(pattern string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			err := fmt.Errorf("http method %s not support", req.Method)
			s.Log.Error(err, "error occur when upload sync options")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			s.Log.Error(err, "read request body error")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// if cancel context is closed, create a new cancel context and sign it to server
		if errors.Is(s.ctx.Err(), context.Canceled) {
			ctx, cancel := context.WithCancel(context.Background())
			s.ctx = ctx
			s.cancel = cancel
		}

		// if terminate signal is not none, then terminate the task in progress,
		// and pop all tasks waiting in event queue,
		jobName := req.URL.Query().Get(common.TerminateSignal)
		if jobName != "" {
			s.cancel()
		}

		fileName := req.Header.Get("filename")
		if fileName == "" {
			e := fmt.Errorf("handle upload %s fail", pattern)
			s.Log.Error(e, "can not get filename from request header")
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if fileName == common.TerminateSignal {
			s.Log.V(1).Info("upload terminate signal successfully")
			w.WriteHeader(http.StatusOK)
			return
		}

		dirPath := s.svrOpt.ServerDir + pattern
		filePath := dirPath + "/" + fileName
		err = ioutil.WriteFile(filePath, body, os.ModePerm)
		if err != nil {
			s.Log.Error(err, "write file error", "file", filePath)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.Log.V(1).Info("upload options success", "file", filePath, "route", pattern)

		w.WriteHeader(http.StatusOK)
		return
	}
}

func (s *Server) watchAndDo() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			s.handleEvent(event)
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			s.Log.Error(err, "watcher get error")
		}
	}
}

func (s *Server) handleEvent(event fsnotify.Event) {
	s.Log.V(1).Info("get event", "event", event.String())
	if event.Op != fsnotify.Create {
		s.Log.V(1).Info("event operation is create")
		return
	}

	switch extractPattern(event.Name) {
	case common.PathSyncOptions:
		s.do(common.PathSyncOptions, event.Name)
	case common.PathClearOptions:
		s.do(common.PathClearOptions, event.Name)
	case common.PathRmrOptions:
		s.do(common.PathRmrOptions, event.Name)
	case common.PathWarmupOptions:
		s.do(common.PathWarmupOptions, event.Name)
	default:
		err := fmt.Errorf("extract pattern error")
		s.Log.Error(err, "can not deal with none")
	}
}

func (s *Server) do(pattern string, optionFile string) {
	body, err := ioutil.ReadFile(optionFile)
	if err != nil {
		msg := fmt.Sprintf("read file %s error: %s", optionFile, err.Error())
		s.writeResultFile(common.JobStatusFail, msg, pattern, optionFile)
		return
	}

	doer, exist := s.doers[pattern]
	if !exist {
		msg := fmt.Sprintf("can not find pattern's doer")
		s.writeResultFile(common.JobStatusFail, msg, pattern, optionFile)
		return
	}
	s.writeResultFile(common.JobStatusRunning, "", pattern, optionFile)

	if err := doer(body); err != nil {
		s.writeResultFile(common.JobStatusFail, err.Error(), pattern, optionFile)
		return
	}

	s.writeResultFile(common.JobStatusSuccess, "", pattern, optionFile)
}

func (s *Server) writeResultFile(status common.JobStatus, message string, pattern string, optionFile string) {
	if message != "" {
		err := errors.New(message)
		s.Log.Error(err, "")
	}
	result := common.JobResult{
		Status:  status,
		Message: message,
	}
	body, err := json.Marshal(result)
	if err != nil {
		s.Log.Error(err, "marshal result error", "result", result)
		return
	}

	resultPath, exist := s.optResMap[pattern]
	if !exist {
		s.Log.Error(errors.New("result pattern not find"), "")
		return
	}

	optionPaths := strings.Split(optionFile, "/")
	filename := optionPaths[len(optionPaths)-1]
	filePath := s.svrOpt.ServerDir + resultPath + "/" + filename

	if err := ioutil.WriteFile(filePath, body, os.ModePerm); err != nil {
		s.Log.Error(err, "write file error", "file", filePath)
	}
	s.Log.V(1).Info("write result success",
		"file", filePath, "result", result, "pattern", pattern)
}

func (s *Server) doSync(body []byte) error {
	s.Log.V(1).Info("begin do sync")
	opt := &v1alpha1.SyncJobOptions{}
	err := json.Unmarshal(body, opt)
	if err != nil {
		return err
	}
	return s.DoSyncJob(s.ctx, opt, s.Log)
}

func (s *Server) doClear(body []byte) error {
	s.Log.V(1).Info("begin do clear")
	opt := &v1alpha1.ClearJobOptions{}
	err := json.Unmarshal(body, opt)
	if err != nil {
		return err
	}
	return s.DoClearJob(s.ctx, opt, s.Log)
}

func (s *Server) doWarmup(body []byte) error {
	s.Log.V(1).Info("begin do warmup")
	opt := &v1alpha1.WarmupJobOptions{}
	err := json.Unmarshal(body, opt)
	if err != nil {
		return err
	}
	return s.DoWarmupJob(s.ctx, opt, s.Log)
}

func (s *Server) doRmr(body []byte) error {
	s.Log.V(1).Info("begin do rmr")
	opt := &v1alpha1.RmrJobOptions{}
	err := json.Unmarshal(body, opt)
	if err != nil {
		return err
	}
	return s.DoRmrJob(s.ctx, opt, s.Log)
}

func (s *Server) addWatchDirs(patterns ...string) error {
	for _, pattern := range patterns {
		err := s.watcher.Add(s.svrOpt.ServerDir + pattern)
		if err != nil {
			return fmt.Errorf("add watcher dir %s error", pattern)
		}
	}
	return nil
}

func (s *Server) addStaticHandlers(patterns ...string) error {
	// Add static file server
	http.Handle("/", http.FileServer(http.Dir(s.svrOpt.ServerDir)))

	for _, pattern := range patterns {
		path := s.svrOpt.ServerDir + pattern
		if _, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			e := os.MkdirAll(path, os.ModePerm)
			if e != nil {
				return e
			}
		}

		handler := http.FileServer(http.Dir(path))
		http.Handle(pattern, http.StripPrefix(pattern, handler))
	}
	return nil
}

func (s *Server) addUploadHandlers(patterns ...string) {
	for _, pattern := range patterns {
		uploadUrl := common.PathUploadPrefix + pattern
		http.HandleFunc(uploadUrl, s.uploadHandleFunc(pattern))
	}
}

func (s *Server) writeCacheStatus() {
	filePath := s.svrOpt.ServerDir +
		common.PathCacheStatus + "/" +
		common.FilePathCacheInfo
	var interval int64 = 5

	for {
		if interval >= s.svrOpt.Interval {
			time.Sleep(time.Duration(s.svrOpt.Interval) * time.Second)
		} else {
			time.Sleep(time.Duration(interval) * time.Second)
			interval += 2
		}

		status := &v1alpha1.CacheStatus{}
		err := s.CreateCacheStatus(s.svrOpt, status)
		if err != nil {
			status.ErrorMassage = err.Error()
			s.Log.Error(err, "")
		}
		body, err := json.Marshal(status)
		if err != nil {
			s.Log.Error(err, "marshal cache status error")
			continue
		}
		err = ioutil.WriteFile(filePath, body, os.ModePerm)
		if err != nil {
			s.Log.Error(err, "write cache status error")
			continue
		}
		s.Log.V(1).Info("write cache status success")
	}
}

func extractPattern(path string) string {
	if strings.Contains(path, common.PathSyncOptions) {
		return common.PathSyncOptions
	}
	if strings.Contains(path, common.PathClearOptions) {
		return common.PathClearOptions
	}
	if strings.Contains(path, common.PathRmrOptions) {
		return common.PathRmrOptions
	}
	if strings.Contains(path, common.PathWarmupOptions) {
		return common.PathWarmupOptions
	}
	return ""
}
