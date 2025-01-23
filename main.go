package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"oldbute.com/bitaxe-metrics/lib"
)

type metrics struct {
	coreVoltage    *prometheus.GaugeVec
	frequency      *prometheus.GaugeVec
	hashRate       *prometheus.GaugeVec
	sharesAccepted *prometheus.GaugeVec
	sharesRejected *prometheus.GaugeVec
	temp           *prometheus.GaugeVec
	vrTemp         *prometheus.GaugeVec
}

func main() {
	reg := prometheus.NewRegistry()

	metrics := NewMetrics(reg)
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
		log.Fatal(http.ListenAndServe(":8077", nil))
	}()

	bitaxes := lib.ScanNetwork()
	log.Printf("Found %d bitaxes", len(bitaxes))

	// Update metrics every x seconds
	metricsPoll := time.NewTicker(10 * time.Second)

	// Scan the network every x minutes (DHCP leases or new Bitaxes)
	scanNetwork := time.NewTicker(10 * time.Minute)

	// Force a scan of the network
	forceScan := make(chan time.Time)

	log.Printf("Starting main loop")
	for {
		select {
		case <-metricsPoll.C:
			log.Printf("Metrics poll tick")
			for _, bitaxe := range bitaxes {
				go func(bitaxe lib.Bitaxe) {
					info := lib.GetSystemInfo(bitaxe.IP)
					if info.Hostname == "" {
						// Bitaxe not responding, force re-scan in 30 seconds
						go func() {
							time.Sleep(30 * time.Second)
							forceScan <- time.Now()
						}()
						log.Printf("Bitaxe %s not responding", bitaxe.Hostname)
						return
					}

					metrics.coreVoltage.WithLabelValues(bitaxe.Hostname).Set(float64(info.CoreVoltage))
					metrics.hashRate.WithLabelValues(bitaxe.Hostname).Set(info.HashRate)
					metrics.frequency.WithLabelValues(bitaxe.Hostname).Set(float64(info.Frequency))
					metrics.sharesAccepted.WithLabelValues(bitaxe.Hostname).Set(float64(info.SharesAccepted))
					metrics.sharesRejected.WithLabelValues(bitaxe.Hostname).Set(float64(info.SharesRejected))
					metrics.temp.WithLabelValues(bitaxe.Hostname).Set(info.Temp)
					metrics.vrTemp.WithLabelValues(bitaxe.Hostname).Set(info.VRTemp)
				}(bitaxe)
			}
		case <-scanNetwork.C:
			log.Printf("Network scan tick")
			bitaxes = lib.ScanNetwork()
			log.Printf("Found %d bitaxes", len(bitaxes))
		case <-forceScan:
			log.Printf("Force scan tick")
			bitaxes = lib.ScanNetwork()
			log.Printf("Found %d bitaxes", len(bitaxes))
		}
	}
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		coreVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "core_voltage",
			Help: "Voltage of the ASICcore in Volts",
		}, []string{"hostname"}),
		frequency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "frequency",
			Help: "Frequency of the ASIC in MHz",
		}, []string{"hostname"}),
		hashRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "hash_rate",
			Help: "Current hash rate as Gh/s",
		}, []string{"hostname"}),
		sharesAccepted: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "shares_accepted",
			Help: "Number of shares accepted",
		}, []string{"hostname"}),
		sharesRejected: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "shares_rejected",
			Help: "Number of shares rejected",
		}, []string{"hostname"}),
		temp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "asic_temperature_celsius",
			Help: "Current temperature of the ASIC.",
		}, []string{"hostname"}),
		vrTemp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vr_temperature_celsius",
			Help: "Current temperature of the voltage regulator.",
		}, []string{"hostname"}),
	}
	reg.MustRegister(m.coreVoltage)
	reg.MustRegister(m.temp)
	reg.MustRegister(m.frequency)
	reg.MustRegister(m.hashRate)
	reg.MustRegister(m.sharesAccepted)
	reg.MustRegister(m.sharesRejected)
	reg.MustRegister(m.vrTemp)
	return m
}
