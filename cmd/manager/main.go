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

package main

import (
	"context"
	"log"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/paddleflow/paddle-operator/api/v1alpha1"
	"github.com/paddleflow/paddle-operator/controllers/extensions/common"
	"github.com/paddleflow/paddle-operator/controllers/extensions/driver"
	"github.com/paddleflow/paddle-operator/controllers/extensions/manager"
)

var (
	rootCmdOptions   common.RootCmdOptions
	serverOptions    common.ServerOptions
	rmrJobOptions    v1alpha1.RmrJobOptions
	syncJobOptions   v1alpha1.SyncJobOptions
	clearJobOptions  v1alpha1.ClearJobOptions
	warmupJobOptions v1alpha1.WarmupJobOptions
)

var rootCmd = &cobra.Command{
	Use:   common.CmdRoot,
	Short: "run server or job command",
}

var serverCmd = &cobra.Command{
	Use:   common.CmdServer,
	Short: "run data management server",
	Run: func(cmd *cobra.Command, args []string) {
		server, err := manager.NewServer(&rootCmdOptions, &serverOptions)
		if err != nil {
			log.Fatalf("create server error: %s", err.Error())
		}
		log.Fatal(server.Run())
	},
}

var syncJobCmd = &cobra.Command{
	Use:   common.CmdSync,
	Short: "sync data or metadata from source to the cache engine",
	Run: func(cmd *cobra.Command, args []string) {
		driverName := v1alpha1.DriverName(rootCmdOptions.Driver)
		d, err := driver.GetDriver(driverName)
		if err != nil {
			log.Fatalf("get driver %s error: %s", driverName, err.Error())
		}
		zapLog := zap.New(func(o *zap.Options) {
			o.Development = rootCmdOptions.Development
		})
		log.Fatal(d.DoSyncJob(context.Background(), &syncJobOptions, zapLog))
	},
}

var warmupJobCmd = &cobra.Command{
	Use:   common.CmdWarmup,
	Short: "warm up data from remote storage to local host",
	Run: func(cmd *cobra.Command, args []string) {
		driverName := v1alpha1.DriverName(rootCmdOptions.Driver)
		d, err := driver.GetDriver(driverName)
		if err != nil {
			log.Fatalf("get driver %s error: %s", driverName, err.Error())
		}
		zapLog := zap.New(func(o *zap.Options) {
			o.Development = rootCmdOptions.Development
		})
		log.Fatal(d.DoWarmupJob(context.Background(), &warmupJobOptions, zapLog))
	},
}

var rmrJobCmd = &cobra.Command{
	Use:   common.CmdRmr,
	Short: "remove data from cache engine storage backend",
	Run: func(cmd *cobra.Command, args []string) {
		driverName := v1alpha1.DriverName(rootCmdOptions.Driver)
		d, err := driver.GetDriver(driverName)
		if err != nil {
			log.Fatalf("get driver %s error: %s", driverName, err.Error())
		}
		zapLog := zap.New(func(o *zap.Options) {
			o.Development = rootCmdOptions.Development
		})
		log.Fatal(d.DoSyncJob(context.Background(), &syncJobOptions, zapLog))
	},
}

var clearJobCmd = &cobra.Command{
	Use:   common.CmdClear,
	Short: "clear cache data from local host",
	Run: func(cmd *cobra.Command, args []string) {
		driverName := v1alpha1.DriverName(rootCmdOptions.Driver)
		d, err := driver.GetDriver(driverName)
		if err != nil {
			log.Fatalf("get driver %s error: %s", driverName, err.Error())
		}
		zapLog := zap.New(func(o *zap.Options) {
			o.Development = rootCmdOptions.Development
		})
		log.Fatal(d.DoClearJob(context.Background(), &clearJobOptions, zapLog))
	},
}

func init() {
	// initialize options for root command
	rootCmd.PersistentFlags().StringVar(&rootCmdOptions.Driver, "driver", string(driver.JuiceFSDriver), "specify the cache engine")
	rootCmd.PersistentFlags().BoolVar(&rootCmdOptions.Development, "development", false, "configures the logger to use a Zap development config")

	// initialize options for runtime server
	serverCmd.Flags().IntVar(&serverOptions.ServerPort, "serverPort", common.RuntimeServicePort, "the port for runtime service")
	serverCmd.Flags().StringVar(&serverOptions.ServerDir, "serverDir", common.PathServerRoot, "the root dir for static file service")
	serverCmd.Flags().StringSliceVar(&serverOptions.CacheDirs, "cacheDirs", nil, "cache data directories that mounted to the container")
	serverCmd.Flags().StringVar(&serverOptions.DataDir, "dataDir", "", "the sample set data path mounted to the container")
	serverCmd.Flags().Int64Var(&serverOptions.Interval, "interval", common.RuntimeCacheInterval, "time interval for writing cache status to specified path")
	serverCmd.Flags().Int64Var(&serverOptions.Timeout, "timeout", common.RuntimeCacheInterval, "The timeout period of the command to collect cached data information")

	// initialize options for sync job command
	syncJobCmd.Flags().StringVar(&syncJobOptions.Source, "source", "", "data source that need sync to cache engine")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Destination, "destination", "", "destination path data should sync to")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Start, "start", "", "the first KEY to sync")
	syncJobCmd.Flags().StringVar(&syncJobOptions.End, "end", "", "the last KEY to sync")
	syncJobCmd.Flags().IntVar(&syncJobOptions.Threads, "threads", 10, "number of concurrent threads")
	syncJobCmd.Flags().IntVar(&syncJobOptions.HttpPort, "http-port", 6070, "HTTP PORT to listen to")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.Update, "update", false, "update existing file if the source is newer")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.ForceUpdate, "force-update", false, "always update existing file")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.Perms, "perms", false, "preserve permissions")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.Dirs, "dirs", false, "Sync directories or holders")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.Dry, "dry", false, "Don't copy file")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.DeleteSrc, "delete-src", false, "delete objects from source after synced")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.DeleteDst, "delete-dst", false, "delete extraneous objects from destination")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Exclude, "exclude", "", "exclude keys containing PATTERN (POSIX regular expressions)")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Include, "include", "", "only include keys containing PATTERN (POSIX regular expressions)")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Manager, "manager", "", "manager address")
	syncJobCmd.Flags().StringVar(&syncJobOptions.Worker, "worker", "", "hosts (seperated by comma) to launch worker")
	syncJobCmd.Flags().IntVar(&syncJobOptions.BWLimit, "bwlimit", 0, "limit bandwidth in Mbps (0 means unlimited)")
	syncJobCmd.Flags().BoolVar(&syncJobOptions.NoHttps, "no-https", false, "do not use HTTPS")

	// initialize options for warmup job command
	warmupJobCmd.Flags().StringSliceVar(&warmupJobOptions.Paths, "paths", nil, "A list of paths need to build cache")
	warmupJobCmd.Flags().Int32Var(&warmupJobOptions.Partitions, "partitions", 0, "the partition number of sampleset")
	warmupJobCmd.Flags().StringVar(&warmupJobOptions.Strategy.Name, "strategyName", "random", "the data warmup strategy name")
	warmupJobCmd.Flags().StringVar(&warmupJobOptions.File, "file", "", "file containing a list of paths")
	warmupJobCmd.Flags().IntVar(&warmupJobOptions.Threads, "threads", 50, "number of concurrent workers")

	// initialize options for rmr job command
	rmrJobCmd.Flags().StringSliceVar(&rmrJobOptions.Paths, "paths", nil, "the data mount paths that need to be remove")

	// initialize options for clear job command
	clearJobCmd.Flags().StringSliceVar(&clearJobOptions.Paths, "paths", nil, "the cache data paths that need to be clear")

	rootCmd.AddCommand(serverCmd, syncJobCmd, warmupJobCmd, rmrJobCmd, clearJobCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
