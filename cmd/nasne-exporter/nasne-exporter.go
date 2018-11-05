package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	flagNasneAddr   = "nasne-addr"
	flagListen      = "listen"
	flagMetricsPath = "metrics-path"
)

func main() {
	c := NewCommand()

	err := c.Execute()
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "nasne-exporter",
		Short: "nasne exporter",
		RunE:  RunNasneExporter,
	}

	cmd.Flags().StringSlice(flagNasneAddr, nil, "Address of Nasne")
	cmd.Flags().String(flagListen, ":8080", "Listen")
	cmd.Flags().String(flagMetricsPath, "/metrics", "Path of metrics")

	flag.Lookup("logtostderr").Value.Set("true")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func RunNasneExporter(cmd *cobra.Command, args []string) error {
	glog.V(2).Info("start nasne-exporter")

	nasneAddr, err := cmd.Flags().GetStringSlice(flagNasneAddr)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagNasneAddr, nasneAddr)

	listen, err := cmd.Flags().GetString(flagListen)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagListen, listen)

	metricsPath, err := cmd.Flags().GetString(flagMetricsPath)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagMetricsPath, metricsPath)

	nc := collector.NewNasneCollector(nasneAddr)
	go nc.Run()

	reg := prometheus.NewRegistry()
	nc.RegisterCollector(reg)

	srv := &http.Server{
		Addr:    listen,
		Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	}

	errCh := make(chan error, 0)
	defer close(errCh)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 0)
	defer close(sigCh)

	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-sigCh:
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if err := srv.Shutdown(ctx); err != nil {
			glog.Error(err)
		}
	case err := <-errCh:
		glog.Error(err)
	}

	return nil
}
