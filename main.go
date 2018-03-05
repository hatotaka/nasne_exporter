package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/hatotaka/nasune-exporter/pkg/nasneclient"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ListenPort = ":8080"
)

func main() {
	m := initMetrics()

	nasneIPs := []string{
		"10.0.1.23",
		"10.0.1.25",
		"10.0.1.22",
	}

	go func() {
		for {
			for _, ip := range nasneIPs {
				client, err := nasneclient.NewNasneClient(ip)
				if err != nil {
					log.Fatal(err)
				}

				bn, err := client.GetBoxName()
				if err != nil {
					log.Fatal(err)
				}

				m.Uptime.With(prometheus.Labels{
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

					m.Info.With(prometheus.Labels{
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

						m.HDDTotal.With(label).Set(hddInfo.HDD.TotalVolumeSize)
						m.HDDUsed.With(label).Set(hddInfo.HDD.UsedVolumeSize)

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

					m.DtcpipClientTotal.With(label).Set(float64(dtcpipClientList.Number))
				}

			}

			time.Sleep(60 * time.Second)
		}
	}()

	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(ListenPort, nil)

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
