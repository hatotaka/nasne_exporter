package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/collector"
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

	// debug
	nasneAddr := []string{
		"10.0.1.23",
		"10.0.1.25",
		"10.0.1.22",
	}

	cmd.Flags().StringSlice(flagNasneAddr, nasneAddr, "Address of Nasne")
	cmd.Flags().String(flagListen, ":8080", "Listen")
	cmd.Flags().String(flagMetricsPath, "/metrics", "Path of metrics")

	flag.Lookup("logtostderr").Value.Set("true")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func RunNasneExporter(cmd *cobra.Command, args []string) error {
	flag.Parse()
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

	collector := collector.NewNasneCollector(nasneAddr)
	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	go collector.Run()

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
