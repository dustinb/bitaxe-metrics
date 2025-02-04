package lib

import (
	"log"
	"net/http"
	"sync"

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
	sharesAccepted     *prometheus.CounterVec
	sharesRejected     *prometheus.CounterVec
	temp               *prometheus.GaugeVec
	vrTemp             *prometheus.GaugeVec
}

var metrics *prometheusMetrics

// Store previous shares accepted and rejected for each Bitaxe
var sharesMutex sync.Mutex
var sharesAcceptedLast map[string]int
var sharesRejectedLast map[string]int

func StartMetrics() {
	reg := prometheus.NewRegistry()

	metrics = newMetrics(reg)

	sharesAcceptedLast = make(map[string]int)
	sharesRejectedLast = make(map[string]int)

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
	metrics.temp.WithLabelValues(hostname).Set(info.Temp)
	metrics.vrTemp.WithLabelValues(hostname).Set(info.VRTemp)

	sharesMutex.Lock()
	defer sharesMutex.Unlock()

	if _, ok := sharesAcceptedLast[info.MacAddr]; !ok {
		sharesAcceptedLast[info.MacAddr] = info.SharesAccepted
	} else {
		for i := 0; i < info.SharesAccepted-sharesAcceptedLast[info.MacAddr]; i++ {
			metrics.sharesAccepted.WithLabelValues(hostname).Inc()
		}
		sharesAcceptedLast[info.MacAddr] = info.SharesAccepted
	}

	if _, ok := sharesRejectedLast[info.MacAddr]; !ok {
		sharesRejectedLast[info.MacAddr] = info.SharesRejected
	} else {
		for i := 0; i < info.SharesRejected-sharesRejectedLast[info.MacAddr]; i++ {
			metrics.sharesRejected.WithLabelValues(hostname).Inc()
		}
		sharesRejectedLast[info.MacAddr] = info.SharesRejected
	}

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
		sharesAccepted: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "shares_accepted",
			Help: "Number of shares accepted",
		}, []string{"hostname"}),
		sharesRejected: prometheus.NewCounterVec(prometheus.CounterOpts{
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
