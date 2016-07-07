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

	"github.com/tarm/serial"
)

// Data retrieve from Plantower PMS 5003 air quality sensor
type pms5003Data struct {
	pm1       int
	pm25      int
	pm10      int
	timestamp time.Time
}

func readPmsData() *pms5003Data {
	c := &serial.Config{Name: "/dev/pms5003", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer s.Close()

	buf := make([]byte, 128)
	var serialData []byte

	// wait until second frame
	for i := 0; i < 2; i++ {
		// could start reading in the middle of a frame, need to wait until we capture the beginning of a data frame
		var n int
		for len(buf) == 0 && buf[0] != 0x42 {
			n, err = s.Read(buf)
			if err != nil {
				log.Println(err)
				return nil
			}
			//log.Printf(hex.Dump(buf[:n]))
		}

		serialData = make([]byte, n)
		copy(serialData, buf[:n])

		//log.Printf("%d", len(data))

		// need to capture a full 32 bytes
		for len(serialData) < 32 {
			n, err := s.Read(buf)
			if err != nil {
				log.Println(err)
				return nil
			}
			serialData = append(serialData, buf[:n]...)
		}
	}

	// log.Printf(hex.Dump(data))

	data := &pms5003Data{
		pm1:       int(serialData[4])<<8 | int(serialData[5]),
		pm25:      int(serialData[6])<<8 | int(serialData[7]),
		pm10:      int(serialData[8])<<8 | int(serialData[9]),
		timestamp: time.Now(),
	}

	log.Printf("PMS %d %d %d %v", data.pm1, data.pm25, data.pm10, data.timestamp)

	return data
}

// Data retrieved from SDS 018 air quality sensor
type sds018Data struct {
	pm25      float32
	pm10      float32
	timestamp time.Time
}

func readSdsData() *sds018Data {
	c := &serial.Config{Name: "/dev/sds018", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	buf := make([]byte, 128)

	for i := 0; i < 2; i++ {
		_, err = s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
	}

	data := &sds018Data{
		pm25:      float32(int(buf[3])<<8|int(buf[2])) / 10.0,
		pm10:      float32(int(buf[5])<<8|int(buf[4])) / 10.0,
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
	c := &serial.Config{Name: "/dev/dht22", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer s.Close()

	reader := bufio.NewReader(s)

	// use second line, since first might have been truncated
	var stringData string
	for i := 0; i < 2; i++ {
		stringData, err = reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return nil
		}
	}

	data := &dht22Data{
		timestamp: time.Now(),
	}
	n, err := fmt.Sscanf(stringData, "%f %f", &data.temperature, &data.humidity)
	if err != nil || n != 2 {
		log.Printf("Unable to parse %s", stringData)
		return nil
	}

	log.Printf("DHT %v %v %v", data.temperature, data.humidity, data.timestamp)

	return data
}

// JSONValue exists for json marshalling
type JSONValue struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float32   `json:"value"`
}

// JSONValues exists for json marshalling
type JSONValues struct {
	Pms5003pm1  []JSONValue `json:"PMS5003_PM1,omitempty"`
	Pms5003pm25 []JSONValue `json:"PMS5003_PM25,omitempty"`
	Pms5003pm10 []JSONValue `json:"PMS5003_PM10,omitempty"`
	Sds018pm25  []JSONValue `json:"SDS018_PM25,omitempty"`
	Sds018pm10  []JSONValue `json:"SDS018_PM10,omitempty"`
	Dht22Temp   []JSONValue `json:"DHT22_Temperature,omitempty"`
	Dht22Humi   []JSONValue `json:"DHT22_Humidity,omitempty"`
}

// JSONWrapper exists for json marshalling
type JSONWrapper struct {
	Values JSONValues `json:"values"`
}

func constructJSON(pmsData *pms5003Data, sdsData *sds018Data, dhtData *dht22Data) string {

	jsonPackage := &JSONWrapper{
		Values: JSONValues{},
	}

	if pmsData != nil {
		jsonPackage.Values.Pms5003pm1 = []JSONValue{
			JSONValue{Timestamp: pmsData.timestamp, Value: float32(pmsData.pm1)},
		}
		jsonPackage.Values.Pms5003pm25 = []JSONValue{
			JSONValue{Timestamp: pmsData.timestamp, Value: float32(pmsData.pm25)},
		}
		jsonPackage.Values.Pms5003pm10 = []JSONValue{
			JSONValue{Timestamp: pmsData.timestamp, Value: float32(pmsData.pm10)},
		}
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

	result, err := json.Marshal(jsonPackage)
	if err != nil {
		log.Fatal(err)
	}

	return string(result)
}

func main() {

	// Read data
	var pmsData *pms5003Data
	var sdsData *sds018Data
	var dhtData *dht22Data

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		pmsData = readPmsData()
	}()
	go func() {
		defer wg.Done()
		sdsData = readSdsData()
	}()
	go func() {
		defer wg.Done()
		dhtData = readDhtData()
	}()

	wg.Wait()
	if pmsData == nil && sdsData == nil && dhtData == nil {
		log.Fatal("Failed to read any data")
	}

	m2xKey, found := os.LookupEnv("XM2XKEY")
	if !found {
		log.Fatal("Environment variable X-M2X-KEY not found")
	}
	url := "http://api-m2x.att.com/v2/devices/4f62aee459c385ac304f53bc16a26bad/updates"

	jsonBody := constructJSON(pmsData, sdsData, dhtData)
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
