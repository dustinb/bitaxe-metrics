package lib

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

const (
	datasourceUID = "cebm7402f3ncwf" // Should match the UID in your datasource.yml
)

// Host struct to hold discovered hosts
type host struct {
	Name string
}

// generatePanel creates a Grafana panel for a given host
func generatePanel(hostName string, x, y int) map[string]interface{} {
	return map[string]interface{}{
		"title": fmt.Sprintf("Hash Rate - %s", hostName),
		"type":  "timeseries",
		"gridPos": map[string]int{
			"x": x, "y": y, "w": 12, "h": 6,
		},
		"datasource": map[string]string{
			"type": "prometheus",
			"uid":  datasourceUID,
		},
		"targets": []map[string]interface{}{
			{
				"expr":         fmt.Sprintf(`avg_over_time(hash_rate{hostname="%s"}[20m])`, hostName),
				"legendFormat": "Hash Rate Gh/s",
				"refId":        "A",
			},
			{
				"expr":         fmt.Sprintf(`expected_hash_rate{hostname="%s"}`, hostName),
				"legendFormat": "Expected",
				"refId":        "B",
			},
		},
	}
}

// generateDashboard creates a dashboard JSON dynamically
func generateDashboard(hosts []host) ([]byte, error) {
	var panels []map[string]interface{}
	x, y := 0, 0

	for i, host := range hosts {
		panels = append(panels, generatePanel(host.Name, x, y))
		if (i+1)%2 == 0 { // Place panels in a grid
			x = 0
			y += 6
		} else {
			x = 12
		}
	}

	dashboard := map[string]interface{}{
		"title":         "Hash Rate",
		"uid":           uuid.New().String(),
		"schemaVersion": 40,
		"time": map[string]string{
			"from": "now-6h",
			"to":   "now",
		},
		"panels": panels,
	}

	return json.Marshal(dashboard)
}

func CreateDashboard(bitaxes []Bitaxe) {
	// Check if the dashboard file exists
	if _, err := os.Stat("./grafana/dashboards/hash_rate.json"); !os.IsNotExist(err) {
		fmt.Println("Dashboard file already exists, skipping creation...")
		return
	}

	hosts := []host{}
	for _, bitaxe := range bitaxes {
		hosts = append(hosts, host{Name: bitaxe.Hostname})
	}

	dashboardJSON, err := generateDashboard(hosts)
	if err != nil {
		fmt.Println("Error generating dashboard:", err)
		os.Exit(1)
	}

	err = os.WriteFile("./grafana/dashboards/hash_rate.json", dashboardJSON, 0644)
	if err != nil {
		fmt.Println("Error writing dashboard file:", err)
		os.Exit(1)
	}

	fmt.Printf("Dashboard JSON: %s", string(dashboardJSON))
}
