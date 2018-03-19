package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/nasneclient"
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
	glog.Info("start nasne-exporter")

	nasneAddr, err := cmd.Flags().GetStringSlice(flagNasneAddr)
	if err != nil {
		return err
	}
	glog.Infof("%v = %v", flagNasneAddr, nasneAddr)

	listen, err := cmd.Flags().GetString(flagListen)
	if err != nil {
		return err
	}
	glog.Infof("%v = %v", flagListen, nasneAddr)

	metricsPath, err := cmd.Flags().GetString(flagMetricsPath)
	if err != nil {
		return err
	}
	glog.Infof("%v = %v", flagMetricsPath, metricsPath)

	nasneExporter, err := NewNasneExporter(nasneAddr, listen)
	if err != nil {
		return err
	}

	nasneExporter.Run()

	srv := &http.Server{
		Addr:    listen,
		Handler: nasneExporter.Handler(),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			glog.Info(err)
		}
	}()

	// シグナルを待つ
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	// シグナルを受け取ったらShutdown
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		log.Print(err)
	}

	/*
		http.Handle(metricsPath, nasneExporter.Handler())
		http.ListenAndServe(listen, nil)
	*/

	return nil
}

type NasneExporter struct {
	nasneAddrs []string
	listen     string
	metrics    MetricInfo
}

func NewNasneExporter(nasneAddrs []string, listen string) (*NasneExporter, error) {
	return &NasneExporter{nasneAddrs: nasneAddrs, listen: listen, metrics: initMetrics()}, nil
}

func (ne *NasneExporter) Handler() http.Handler {
	return promhttp.HandlerFor(ne.metrics.Registry, promhttp.HandlerOpts{})
}

func (ne *NasneExporter) Run() {
	go func() {
		for {
			for _, ip := range ne.nasneAddrs {
				glog.Infof("start (ip = %v)", ip)
				client, err := nasneclient.NewNasneClient(ip)
				if err != nil {
					log.Fatal(err)
				}

				bn, err := client.GetBoxName()
				if err != nil {
					log.Fatal(err)
				}

				ne.metrics.Uptime.With(prometheus.Labels{
					"name":   bn.Name,
					"ipaddr": ip,
				}).Set(1)

				{
					softwareVersion, err := client.GetSoftwareVersion()
					if err != nil {
						log.Fatal(err)
					}

					hardwareVersion, err := client.GetHardwareVersion()
					if err != nil {
						log.Fatal(err)
					}

					ne.metrics.Info.With(prometheus.Labels{
						"name":             bn.Name,
						"software_version": softwareVersion.SoftwareVersion,
						"hardware_version": strconv.Itoa(hardwareVersion.HardwareVersion),
						"product_name":     hardwareVersion.ProductName,
					}).Set(1)
				}

				{
					hddList, err := client.GetHDDList()
					if err != nil {
						log.Fatal(err)
					}

					for _, hdd := range hddList.HDD {
						hddInfo, err := client.GetHDDInfo(hdd.ID)
						if err != nil {
							log.Fatal(err)
						}

						label := prometheus.Labels{
							"name":       bn.Name,
							"id":         strconv.Itoa(hddInfo.HDD.ID),
							"format":     hddInfo.HDD.Format,
							"hdd_name":   hddInfo.HDD.Name,
							"vendor_id":  hddInfo.HDD.VendorID,
							"product_id": hddInfo.HDD.ProductID,
						}

						ne.metrics.HDDTotal.With(label).Set(hddInfo.HDD.TotalVolumeSize)
						ne.metrics.HDDUsed.With(label).Set(hddInfo.HDD.UsedVolumeSize)

					}
				}

				{
					dtcpipClientList, err := client.GetDTCPIPClientList()
					if err != nil {
						log.Fatal(err)
					}

					label := prometheus.Labels{
						"name": bn.Name,
					}

					ne.metrics.DtcpipClientTotal.With(label).Set(float64(dtcpipClientList.Number))
				}
				glog.Infof("end (ip = %v)", ip)
			}

			time.Sleep(60 * time.Second)
		}
	}()
}

func main2() {

}

type MetricInfo struct {
	Registry *prometheus.Registry

	Uptime            *prometheus.GaugeVec
	HDDTotal          *prometheus.GaugeVec
	HDDUsed           *prometheus.GaugeVec
	DtcpipClientTotal *prometheus.GaugeVec
	Info              *prometheus.GaugeVec
}

const (
	namespace = "nasne"
)

func initMetrics() MetricInfo {

	m := MetricInfo{}
	reg := prometheus.NewRegistry()

	m.Registry = reg

	m.Info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "info",
		Help:      "info of nasne",
	},
		[]string{
			"name",
			"software_version",
			"hardware_version",
			"product_name",
		})
	reg.MustRegister(m.Info)

	m.Uptime = prometheus.NewGaugeVec(prometheus.GaugeOpts{

		Namespace: namespace,
		Name:      "uptime",
		Help:      "nasne uptime",
	},
		[]string{
			"name",
			"ipaddr",
		})
	reg.MustRegister(m.Uptime)

	m.HDDTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "hdd_byte_total",
		Help:      "nasne hdd byte total",
	},
		[]string{
			"name",
			"id",
			"format",
			"hdd_name",
			"vendor_id",
			"product_id",
		})
	reg.MustRegister(m.HDDTotal)

	m.HDDUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "hdd_byte_used",
		Help:      "nasne hdd byte used",
	},
		[]string{
			"name",
			"id",
			"format",
			"hdd_name",
			"vendor_id",
			"product_id",
		})
	reg.MustRegister(m.HDDUsed)

	m.DtcpipClientTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "dtcpip_client_total",
		Help:      "nasne dtcpip client total",
	},
		[]string{
			"name",
		})
	reg.MustRegister(m.DtcpipClientTotal)

	return m
}
