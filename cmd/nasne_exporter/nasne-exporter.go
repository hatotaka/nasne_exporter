package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne_exporter/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	flagNasneAddr        = "nasne-addr"
	flagPort             = "port"
	flagMetricsPath      = "metrics-path"
	flagDefaultCollector = "default-collector"
)

func main() {
	flag.CommandLine.Parse([]string{})

	c := NewCommand()

	err := c.Execute()
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "nasne_exporter",
		Short: "nasne exporter",
		RunE:  RunNasneExporter,
	}

	cmd.Flags().StringSlice(flagNasneAddr, nil, "The address list of nasne.")
	cmd.Flags().Int(flagPort, 8080, "The port of the endpoint.")
	cmd.Flags().String(flagMetricsPath, "/metrics", "The path of metrics.")
	cmd.Flags().Bool(flagDefaultCollector, true, "Enable prometheus/client_go default collecter (ProcessCollector and GoCollectora)")

	flag.Lookup("logtostderr").Value.Set("true")
	cmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	return cmd
}

func RunNasneExporter(cmd *cobra.Command, args []string) error {
	glog.V(2).Info("start nasne_exporter")

	nasneAddr, err := cmd.Flags().GetStringSlice(flagNasneAddr)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagNasneAddr, nasneAddr)

	port, err := cmd.Flags().GetInt(flagPort)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagPort, port)

	metricsPath, err := cmd.Flags().GetString(flagMetricsPath)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagMetricsPath, metricsPath)

	defaultCollector, err := cmd.Flags().GetBool(flagDefaultCollector)
	if err != nil {
		return err
	}
	glog.V(2).Infof("%v = %v", flagDefaultCollector, defaultCollector)

	reg := prometheus.NewRegistry()

	nc := collector.NewNasneCollector(nasneAddr)
	go nc.Run()
	nc.RegisterCollectors(reg)

	if defaultCollector {
		reg.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
		reg.MustRegister(prometheus.NewGoCollector())
	}

	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
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

	glog.V(2).Info("stop nasne_exporter")
	return nil
}
