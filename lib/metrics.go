package lib

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type prometheusMetrics struct {
	coreVoltage        *prometheus.GaugeVec
	efficiency         *prometheus.GaugeVec
	expectedEfficiency *prometheus.GaugeVec
	expectedHashRate   *prometheus.GaugeVec
	frequency          *prometheus.GaugeVec
	hashRate           *prometheus.GaugeVec
	power              *prometheus.GaugeVec
	sharesAccepted     *prometheus.GaugeVec
	sharesRejected     *prometheus.GaugeVec
	temp               *prometheus.GaugeVec
	vrTemp             *prometheus.GaugeVec
}

var metrics *prometheusMetrics

func StartMetrics() {
	reg := prometheus.NewRegistry()

	metrics = newMetrics(reg)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	go func() {
		log.Printf("Starting metrics server on port 8077")
		log.Fatal(http.ListenAndServe(":8077", nil))
	}()
}

func Measure(hostname string, info Info) {
	metrics.coreVoltage.WithLabelValues(hostname).Set(float64(info.CoreVoltage))
	metrics.hashRate.WithLabelValues(hostname).Set(info.HashRate)
	metrics.expectedHashRate.WithLabelValues(hostname).Set(ExpectedHashRate(info))
	metrics.efficiency.WithLabelValues(hostname).Set(info.Power / (info.HashRate / 1000))
	metrics.expectedEfficiency.WithLabelValues(hostname).Set(info.Power / (ExpectedHashRate(info) / 1000))
	metrics.frequency.WithLabelValues(hostname).Set(float64(info.Frequency))
	metrics.power.WithLabelValues(hostname).Set(info.Power)
	metrics.sharesAccepted.WithLabelValues(hostname).Set(float64(info.SharesAccepted))
	metrics.sharesRejected.WithLabelValues(hostname).Set(float64(info.SharesRejected))
	metrics.temp.WithLabelValues(hostname).Set(info.Temp)
	metrics.vrTemp.WithLabelValues(hostname).Set(info.VRTemp)
}

func newMetrics(reg prometheus.Registerer) *prometheusMetrics {
	m := &prometheusMetrics{
		coreVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "core_voltage",
			Help: "Voltage of the ASICcore in Volts",
		}, []string{"hostname"}),
		efficiency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "efficiency",
			Help: "Efficiency of the ASIC in J/Th",
		}, []string{"hostname"}),
		expectedEfficiency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "expected_efficiency",
			Help: "Expected efficiency of the ASIC in J/Th",
		}, []string{"hostname"}),
		frequency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "frequency",
			Help: "Frequency of the ASIC in MHz",
		}, []string{"hostname"}),
		hashRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "hash_rate",
			Help: "Current hash rate as Gh/s",
		}, []string{"hostname"}),
		expectedHashRate: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "expected_hash_rate",
			Help: "Expected hash rate as Gh/s",
		}, []string{"hostname"}),
		power: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "power",
			Help: "Power consumption in Watts",
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
	reg.MustRegister(m.efficiency)
	reg.MustRegister(m.expectedEfficiency)
	reg.MustRegister(m.frequency)
	reg.MustRegister(m.hashRate)
	reg.MustRegister(m.expectedHashRate)
	reg.MustRegister(m.power)
	reg.MustRegister(m.sharesAccepted)
	reg.MustRegister(m.sharesRejected)
	reg.MustRegister(m.temp)
	reg.MustRegister(m.vrTemp)
	return m
}
