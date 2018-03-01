package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/solo-io/gloo-storage/crd"
	"github.com/spf13/cobra"

	"github.com/solo-io/gloo-function-discovery/internal/eventloop"
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/gloo/pkg/signals"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var opts bootstrap.Options

var rootCmd = &cobra.Command{
	Use:   "gloo",
	Short: "runs the gloo control plane to manage Envoy as a Function Gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		stop := signals.SetupSignalHandler()
		errs := make(chan error)

		finished := make(chan error)
		go func() { finished <- eventloop.Run(opts, stop, errs) }()
		go func() {
			for {
				select {
				case err := <-errs:
					log.Warnf("discovery error: %v", err)
				}
			}
		}()
		return <-finished
	},
}

func init() {
	// config watcher
	rootCmd.PersistentFlags().StringVar(&opts.ConfigWatcherOptions.Type, "storage.type", bootstrap.WatcherTypeFile, fmt.Sprintf("storage backend for config objects. supported: [%s]", strings.Join(bootstrap.SupportedCwTypes, " | ")))
	rootCmd.PersistentFlags().DurationVar(&opts.ConfigWatcherOptions.SyncFrequency, "storage.refreshrate", time.Second, "refresh rate for polling config")

	// storage watcher
	rootCmd.PersistentFlags().StringVar(&opts.SecretWatcherOptions.Type, "secrets.type", bootstrap.WatcherTypeFile, fmt.Sprintf("storage backend for secrets. supported: [%s]", strings.Join(bootstrap.SupportedSwTypes, " | ")))
	rootCmd.PersistentFlags().DurationVar(&opts.SecretWatcherOptions.SyncFrequency, "secrets.refreshrate", time.Second, "refresh rate for polling secrets")

	// xds port
	rootCmd.PersistentFlags().IntVar(&opts.XdsOptions.Port, "xds.port", 8081, "port to serve envoy xDS services. this port should be specified in your envoy's static config")

	// file
	rootCmd.PersistentFlags().StringVar(&opts.FileOptions.ConfigDir, "file.config.dir", "_gloo_config", "root directory to use for storing gloo config files")
	rootCmd.PersistentFlags().StringVar(&opts.FileOptions.SecretDir, "file.secret.dir", "_gloo_secrets", "root directory to use for storing gloo secret files")

	// kube
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.MasterURL, "master", "", "url of the kubernetes apiserver. not needed if running in-cluster")
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.KubeConfig, "kubeconfig", "", "path to kubeconfig file. not needed if running in-cluster")
	rootCmd.PersistentFlags().StringVar(&opts.KubeOptions.Namespace, "kube.namespace", crd.GlooDefaultNamespace, "namespace to read/write gloo storage objects")

	// vault
	rootCmd.PersistentFlags().StringVar(&opts.VaultOptions.VaultAddr, "vault.addr", "", "url for vault server")
	rootCmd.PersistentFlags().StringVar(&opts.VaultOptions.AuthToken, "vault.token", "", "auth token for reading vault secrets")
	rootCmd.PersistentFlags().IntVar(&opts.VaultOptions.Retries, "vault.retries", 3, "number of times to retry failed requests to vault")
}
