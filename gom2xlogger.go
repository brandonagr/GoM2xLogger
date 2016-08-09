package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"math"

	"github.com/tarm/serial"
	"github.com/cocoonlife/goalsa"
)

// #cgo LDFLAGS: -ldht
// #include "/home/pi/Adafruit_Python_DHT/source/Raspberry_Pi_2/pi_2_dht_read.h"
import "C"

// Data retrieved from SDS 018 air quality sensor
type sds021Data struct {
	pm25      float32
	pm10      float32
	timestamp time.Time
}

func readSdsData(averageOver time.Duration) *sds021Data {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: time.Second * 10}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Print(err)
		return nil
	}
	defer s.Close()

	// discard first reading
	_, err = s.Read(buf)

	var accumulator sds021Data
	count := 0.0

	buf := make([]byte, 128)
	startTime := time.Now()
	for (time.Now() - startTime) < averageOver {
		_, err = s.Read(buf)
		if err != nil {
			log.Print(err)
			return nil
		}

		accumulator.pm25 += float32(int(buf[3])<<8|int(buf[2])) / 10.0
		accumulator.pm10 += float32(int(buf[5])<<8|int(buf[4])) / 10.0
		count += 1.0
	}

	data := &sds018Data{
		pm25:      accumulator.pm25 / count,
		pm10:      accumulator.pm10 / count
		timestamp: time.Now(),
	}

	log.Printf("SDS %v %v %v", data.pm25, data.pm10, data.timestamp)

	return data
}

type dht22Data struct {
	temperature float32
	humidity    float32
	timestamp   time.Time
}

func readDhtData() *dht22Data {

	var humidity C.float
	var temperature C.float

	// discard first reading
	C.pi_2_dht_read(22, 4, &humidity, &temperature)

	C.pi_2_dht_read(22, 4, &humidity, &temperature)

	data := &dht22Data{
		temperature: ((float32)temperature) * 1.8 + 32.0, // C to F
		humidity: (float32)humidity,
		timestamp: time.Now(),
	}

	log.Printf("DHT %v %v %v", data.temperature, data.humidity, data.timestamp)

	return data
}

type soundData struct {
	decibels float32
	timestamp   time.Time
}

func readSoundData(averageOver time.Duration) *soundData {

	c, err := alsa.NewCaptureDevice("hw:1,0", 1, alsa.FormatS16LE, 44100, a$
	if err != nil {
		log.Print(err)
		return nil
	}
	defer c.Close()

	buffer := make([]int16, 8000)

	var averageDb float64
	averageDb = 0.0
	var count int64

	startTime := time.Now()
	for (time.Now() - startTime) < averageOver {
		count, err := c.Read(buffer)

		if count == 0 {
			continue
		}

		for _, value := range buffer[:count] {
			averageDb += 20 * math.Log10(math.Pow((float64)value, 2.0))
		}
	}

	data := &soundData{
		decibels: averageDb / (float64)count,
		timestamp: time.Now(),
	}

	log.Printf("Sound %v %v", data.decibels, data.timestamp)

	return data
}

// JSONValue exists for json marshalling
type JSONValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float32   `json:"value"`
}

// JSONValues exists for json marshalling
type JSONValues struct {
	Sds021pm25  []JSONValue `json:"SDS021_PM25,omitempty"`
	Sds021pm10  []JSONValue `json:"SDS021_PM10,omitempty"`
	Dht22Temp   []JSONValue `json:"DHT22_Temperature,omitempty"`
	Dht22Humi   []JSONValue `json:"DHT22_Humidity,omitempty"`
	Decibels  []JSONValue `json:"Decibels,omitempty"`
}

// JSONWrapper exists for json marshalling
type JSONWrapper struct {
	Values JSONValues `json:"values"`
}

func constructJSON(sdsData *sds021Data, dhtData *dht22Data, sound *soundData) string {

	jsonPackage := &JSONWrapper{
		Values: JSONValues{},
	}

	if sdsData != nil {
		jsonPackage.Values.Sds018pm25 = []JSONValue{
			JSONValue{Timestamp: sdsData.timestamp, Value: sdsData.pm25},
		}
		jsonPackage.Values.Sds018pm10 = []JSONValue{
			JSONValue{Timestamp: sdsData.timestamp, Value: sdsData.pm10},
		}
	}

	if dhtData != nil {
		jsonPackage.Values.Dht22Temp = []JSONValue{
			JSONValue{Timestamp: dhtData.timestamp, Value: dhtData.temperature},
		}
		jsonPackage.Values.Dht22Humi = []JSONValue{
			JSONValue{Timestamp: dhtData.timestamp, Value: dhtData.humidity},
		}
	}

	if sound != nil {
		jsonPackage.Values.Decibels = []JSONValue{
			JSONValue{TimeStamp: sound.timeStamp, Value: sound.decibels}
		}
	}

	result, err := json.Marshal(jsonPackage)
	if err != nil {
		log.Fatal(err)
	}

	return string(result)
}

func main() {

	averageOver := 5 * time.Second

	// Read data
	var sdsData *sds018Data
	var dhtData *dht22Data
	var sound *soundData

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		pmsData = readSoundData(averageOver)
	}()
	go func() {
		defer wg.Done()
		sdsData = readSdsData(averageOver)
	}()
	go func() {
		defer wg.Done()
		dhtData = readDhtData()
	}()

	wg.Wait()
	if pmsData == nil && sdsData == nil && dhtData == nil {
		log.Fatal("Failed to read any data")
	}

	m2xDevice, found := os.LookupEnv("XM2XDEVICE")
	if !found {
		log.Fatal("Environment variable XM2XDEVICE not found")
	}
	m2xKey, found := os.LookupEnv("XM2XKEY")
	if !found {
		log.Fatal("Environment variable XM2XKEY not found")
	}
	url := fmt.Sprint("http://api-m2x.att.com/v2/devices/%s/updates", m2xDevice)

	jsonBody := constructJSON(sdsData, dhtData, sound)
	log.Println(jsonBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(jsonBody))
	req.Header.Set("X-M2X-KEY", m2xKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
