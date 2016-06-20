package main

import (
        "log"
	"encoding/hex"
        "github.com/tarm/serial"
)

func main() {
        c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
        s, err := serial.OpenPort(c)
        if err != nil {
                log.Fatal(err)
        }

//        _, err = s.Write([]byte("test"))
//        if err != nil {
//                log.Fatal(err)
//        }

        //buf := make([]byte, 128)
//for {
//		n, err := s.Read(buf)
//		if err != nil {
//				log.Fatal(err)
//		}
//		log.Printf("%d %d %d", n, int(buf[6]) << 8 | int(buf[7]), int(buf[8]) << 8 |int(buf[9]))
//}

	buf := make([]byte, 128)

	for {
		n, err := s.Read(buf)
		if err != nil {
				log.Fatal(err)
		}
		//log.Printf(hex.Dump(buf[:n]))
		
		if (buf[0] != 0x42) {
			//log.Printf("Not start %d", buf[0])
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])			

		//log.Printf("%d", len(data))
						
		for len(data) < 32 {
			n, err := s.Read(buf)
			//log.Printf("Appending %d", n)
			//log.Printf(hex.Dump(buf[:n]))
			//log.Printf("")
			if err != nil {
					log.Fatal(err)
			}
			//log.Printf("BEFORE")
			//log.Printf(hex.Dump(data))
			data = append(data, buf[:n]...)	
			//log.Printf("AFTER")
			//log.Printf(hex.Dump(data))
			//log.Printf("")
		}

		//log.Printf("FINAL")
		log.Printf(hex.Dump(data))

		log.Printf("%d %d %d", len(data), int(data[4]) << 8 | int(data[5]), int(data[6]) << 8 |int(data[7]))
                log.Printf("%d %d", int(data[10]) << 8 | int(data[11]), int(data[12]) << 8 |int(data[13]))

	}
}
