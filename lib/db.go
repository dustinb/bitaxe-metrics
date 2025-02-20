package lib

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	UpdatedAt   time.Time
	Hostname    string
	MacAddr     string `gorm:"primaryKey"`
	Frequency   int    `gorm:"primaryKey"`
	CoreVoltage int    `gorm:"primaryKey"`
	HashRate    float64
	Efficiency  float64
	Temp        float64
	H           float64
	E           float64
	T           float64
	Count       float64
}

// Unique key for collecting averages
type ConfigKey struct {
	MacAddr     string
	Frequency   int
	CoreVoltage int
}

var db *gorm.DB

func InitDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("config.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Config{})
}

func StoreAverages(averages map[ConfigKey]Config) {
	for key, average := range averages {
		average.UpdatedAt = time.Now()
		average.MacAddr = key.MacAddr
		average.Frequency = key.Frequency
		average.CoreVoltage = key.CoreVoltage
		average.T = average.T / average.Count
		average.H = average.H / average.Count
		average.E = average.E / average.Count
		average.Efficiency = average.Efficiency / average.Count
		average.HashRate = average.HashRate / average.Count
		average.Temp = average.Temp / average.Count
		db.Save(&average)
	}
}
