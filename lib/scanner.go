package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Info struct {
	ASICModel         string  `json:"ASICModel"`
	AutoFanSpeed      bool    `json:"autoFanSpeed"`
	AsicCount         int     `json:"asicCount"`
	BestDifficulty    string  `json:"bestDifficulty"`
	BoardVersion      string  `json:"boardVersion"`
	CoreVoltage       int     `json:"coreVoltage"`
	CoreVoltageActual int     `json:"coreVoltageActual"`
	Current           float64 `json:"current"`
	FanRPM            int     `json:"fanrpm"`
	FanSpeed          int     `json:"fanspeed"`
	FreeHeap          int     `json:"freeHeap"`
	Frequency         int     `json:"frequency"`
	HashRate          float64 `json:"hashRate"`
	Hostname          string  `json:"hostname"`
	MacAddr           string  `json:"macAddr"`
	OverHeatMode      bool    `json:"overheat_mode"`
	Power             float64 `json:"power"`
	SharesAccepted    int     `json:"sharesAccepted"`
	SharesRejected    int     `json:"sharesRejected"`
	SmallCoreCount    int     `json:"smallCoreCount"`
	StratumDifficulty int     `json:"stratumDifficulty"`
	Temp              float64 `json:"temp"`
	UpTimeSeconds     int     `json:"uptimeSeconds"`
	Voltage           float64 `json:"voltage"`
	VRTemp            float64 `json:"vrTemp"`
}

type Bitaxe struct {
	IP       string
	Hostname string
	MacAddr  string
}

const DHCP_START = 1
const DHCP_END = 254

func GetSystemInfo(ip string) Info {
	client := http.Client{}
	client.Timeout = 2 * time.Second
	resp, err := client.Get("http://" + ip + "/api/system/info")
	if err != nil {
		return Info{}
	}
	body, _ := io.ReadAll(resp.Body)
	info := Info{}
	json.Unmarshal(body, &info)
	return info
}

// ScanNetwork scans the network for the Bitaxe
func ScanNetwork() []Bitaxe {
	var bitaxes = make(map[string]Bitaxe)
	var mutex sync.Mutex
	var waitgroup sync.WaitGroup

	log.Printf("Scanning network for Bitaxes...")
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		host, _ := address.(*net.IPNet)
		if host.IP.IsLoopback() {
			continue
		}
		// Check for IPv4
		if host.IP.To4() == nil {
			continue
		}
		log.Print(host.IP.String())
		octets := strings.Split(host.IP.String(), ".")
		network := octets[0] + "." + octets[1] + "." + octets[2] + ".%d"

		for i := DHCP_START; i < DHCP_END+1; i++ {
			ip := fmt.Sprintf(network, i)
			waitgroup.Add(1)
			go func() {
				defer waitgroup.Done()
				info := GetSystemInfo(ip)
				if info.Hostname != "" {
					log.Printf("Found Bitaxe: %s %s", ip, info.Hostname)
					bitaxe := Bitaxe{IP: ip, Hostname: info.Hostname, MacAddr: info.MacAddr}
					mutex.Lock()
					bitaxes[bitaxe.IP] = bitaxe
					mutex.Unlock()
				}
			}()
		}
	}
	waitgroup.Wait()
	var result []Bitaxe
	for _, bitaxe := range bitaxes {
		result = append(result, bitaxe)
	}
	return result
}
