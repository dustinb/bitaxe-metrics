package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"oldbute.com/bitaxe-metrics/lib"
)

const metricsPollInterval = 10 * time.Second
const scanNetworkInterval = 10 * time.Minute
const storeConfigsInterval = 30 * time.Second

func main() {

	lib.StartMetrics()
	lib.InitDB()
	bitaxes := lib.ScanNetwork()
	log.Printf("Found %d bitaxes", len(bitaxes))

	// Create the Hash Rate dashboard
	lib.CreateDashboard(bitaxes)

	// Update metrics every x seconds
	metricsPoll := time.NewTicker(metricsPollInterval)

	// Scan the network every x minutes (DHCP leases or new/removed Bitaxes)
	scanNetwork := time.NewTicker(scanNetworkInterval)

	// Store the configs every x seconds
	storeConfigs := time.NewTicker(storeConfigsInterval)

	// Force a scan of the network
	forceScan := make(chan time.Time)

	// Track averages over time
	var avgMutex sync.Mutex
	var averages = make(map[lib.ConfigKey]lib.Config)

	log.Printf("Starting main loop")
	for {
		select {
		case <-metricsPoll.C:
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

					// Ramp up measurements
					if info.Temp < 1 || info.HashRate < 1 {
						return
					}

					// Store the metrics
					lib.Measure(bitaxe.Hostname, info)

					// Calculate the score
					T := lib.TemperatureScore(info)
					H := lib.HashRateScore(info)
					E := lib.EfficiencyScore(info)

					// Store config/averages
					configKey := lib.ConfigKey{
						MacAddr:     info.MacAddr,
						Frequency:   info.Frequency,
						CoreVoltage: info.CoreVoltage,
					}

					// Update the averages
					avgMutex.Lock()
					if _, ok := averages[configKey]; !ok {
						averages[configKey] = lib.Config{}
					}

					average := averages[configKey]
					average.Hostname = bitaxe.Hostname
					average.T += T
					average.H += H
					average.E += E
					average.Efficiency += info.Power / (info.HashRate / 1000)
					average.HashRate += info.HashRate
					average.Temp += info.Temp
					average.Count++
					averages[configKey] = average
					avgMutex.Unlock()

					fmt.Printf("%s: %dMHz %d\n", bitaxe.Hostname, info.Frequency, info.CoreVoltage)
					fmt.Printf("Score: T: %f H: %f E: %f\n", T, H, E)
					fmt.Printf("Value: T: %f H: %f E: %f\n", info.Temp, info.HashRate, info.Power/(info.HashRate/1000))
					fmt.Printf("Avg  : T: %f H: %f E: %f\n\n", average.Temp/average.Count, average.HashRate/average.Count, average.Efficiency/average.Count)

				}(bitaxe)
			}
		case <-scanNetwork.C:
			metricsPoll.Stop()
			log.Printf("Network scan tick")
			bitaxes = lib.ScanNetwork()
			log.Printf("Found %d bitaxes", len(bitaxes))
			metricsPoll = time.NewTicker(metricsPollInterval)
		case <-forceScan:
			metricsPoll.Stop()
			log.Printf("Force scan tick")
			bitaxes = lib.ScanNetwork()
			log.Printf("Found %d bitaxes", len(bitaxes))
			metricsPoll = time.NewTicker(metricsPollInterval)
		case <-storeConfigs.C:
			lib.StoreAverages(averages)
		}
	}
}
