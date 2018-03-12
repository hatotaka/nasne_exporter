package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/hatotaka/nasne-exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

const (
	ListenPort = ":8080"
)

const (
	flagNasneAddr = "nasne-addr"
	flagListen    = "listen"
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

	cmd.Flags().StringSlice(flagNasneAddr, []string{}, "Address of Nasne")
	cmd.Flags().String(flagListen, ":8080", "Listen")

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
	// debug
	nasneAddr = []string{
		"10.0.1.23",
		"10.0.1.25",
		"10.0.1.22",
	}
	glog.Infof("%s = %v", flagNasneAddr, nasneAddr)

	listen, err := cmd.Flags().GetString(flagListen)
	if err != nil {
		return err
	}
	glog.Infof("%v = %v", flagListen, nasneAddr)

	nasneExporter, err := NewNasneExporter(nasneAddr, listen)

	nasneExporter.Run()
	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(ListenPort, nil)

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

func (ne *NasneExporter) Run() {
	go func() {
		for {
			for _, ip := range ne.nasneAddrs {
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

			}

			time.Sleep(60 * time.Second)
		}
	}()
}

func main2() {

}

type MetricInfo struct {
	Uptime            *prometheus.GaugeVec
	HDDTotal          *prometheus.GaugeVec
	HDDUsed           *prometheus.GaugeVec
	DtcpipClientTotal *prometheus.GaugeVec
	Info              *prometheus.GaugeVec
}

func initMetrics() MetricInfo {

	m := MetricInfo{}

	m.Info = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nasne_info",
		Help: "info of nasne",
	},
		[]string{
			"name",
			"software_version",
			"hardware_version",
			"product_name",
		})
	prometheus.MustRegister(m.Info)

	m.Uptime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nasne_uptime",
		Help: "nasne uptime",
	},
		[]string{
			"name",
			"ipaddr",
		})
	prometheus.MustRegister(m.Uptime)

	m.HDDTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nasne_hdd_byte_total",
		Help: "nasne hdd byte total",
	},
		[]string{
			"name",
			"id",
			"format",
			"hdd_name",
			"vendor_id",
			"product_id",
		})
	prometheus.MustRegister(m.HDDTotal)

	m.HDDUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nasne_hdd_byte_used",
		Help: "nasne hdd byte used",
	},
		[]string{
			"name",
			"id",
			"format",
			"hdd_name",
			"vendor_id",
			"product_id",
		})
	prometheus.MustRegister(m.HDDUsed)

	m.DtcpipClientTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nasne_dtcpip_client_total",
		Help: "nasne dtcpip client total",
	},
		[]string{
			"name",
		})
	prometheus.MustRegister(m.DtcpipClientTotal)

	return m
}
