package main

import (
	"bufio"
	"fmt"
	"log"
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

func main() {

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
}
