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

func main() {

	lib.StartMetrics()

	bitaxes := lib.ScanNetwork()
	log.Printf("Found %d bitaxes", len(bitaxes))

	// Create the Hash Rate dashboard
	lib.CreateDashboard(bitaxes)

	// Update metrics every x seconds
	metricsPoll := time.NewTicker(metricsPollInterval)

	// Scan the network every x minutes (DHCP leases or new/removed Bitaxes)
	scanNetwork := time.NewTicker(scanNetworkInterval)

	// Force a scan of the network
	forceScan := make(chan time.Time)

	// Track averages over time
	var avgMutex sync.Mutex
	var averages = make(map[string]lib.Average)
	for _, bitaxe := range bitaxes {
		averages[bitaxe.MacAddr] = lib.Average{}
	}

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

					// Store the metrics
					lib.Measure(bitaxe.Hostname, info)

					// Calculate the score
					T := lib.TemperatureScore(info)
					H := lib.HashRateScore(info)
					E := lib.EfficiencyScore(info)
					score := T*0.1 + H*0.6 + E*0.3

					// Update the averages
					avgMutex.Lock()
					average := averages[bitaxe.MacAddr]
					average.Score += score
					average.Efficiency += info.Power / (info.HashRate / 1000)
					average.HashRate += info.HashRate
					average.Temp += info.Temp
					average.Count++
					averages[bitaxe.MacAddr] = average
					avgMutex.Unlock()

					fmt.Printf("%s: %dMHz %d\n", bitaxe.Hostname, info.Frequency, info.CoreVoltage)
					fmt.Printf("Score: T: %f H: %f E: %f\n", T, H, E)
					fmt.Printf("Value: T: %f H: %f E: %f\n", info.Temp, info.HashRate, info.Power/(info.HashRate/1000))
					fmt.Printf("Avg  : T: %f H: %f E: %f S: %f\n\n", average.Temp/average.Count, average.HashRate/average.Count, average.Efficiency/average.Count, average.Score/average.Count)

				}(bitaxe)
			}
		case <-scanNetwork.C:
			metricsPoll.Stop()
			log.Printf("Network scan tick")
			bitaxes = lib.ScanNetwork()

			for _, bitaxe := range bitaxes {
				if _, ok := averages[bitaxe.MacAddr]; !ok {
					averages[bitaxe.MacAddr] = lib.Average{}
				}
			}

			log.Printf("Found %d bitaxes", len(bitaxes))
			metricsPoll = time.NewTicker(metricsPollInterval)
		case <-forceScan:
			metricsPoll.Stop()
			log.Printf("Force scan tick")
			bitaxes = lib.ScanNetwork()
			log.Printf("Found %d bitaxes", len(bitaxes))
			metricsPoll = time.NewTicker(metricsPollInterval)
		}
	}
}
