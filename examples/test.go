package main

import (
	"github.com/ziutek/avt414"
	"log"
	"time"
)

func main() {
	a, err := avt414.Open("/dev/ttyUSB0")
	if err != nil {
		log.Fatal(err)
	}

	port := byte('D')

	if err = a.Write(port, 0); err != nil {
		log.Fatal(err)
	}
	if err = a.Setup(port, 0); err != nil {
		log.Fatal(err)
	}
	var i byte
	for {
		if i == 0 {
			i = 0x80
		}
		if err = a.Write(port, i); err != nil {
			log.Fatal(err)
		}
		v, err := a.Read('B')
		if err != nil {
			log.Fatal(err)
		}
		a, err := a.ADC(0)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%x, %.3f", v, 2.5 * float32(a) / 1023)
		time.Sleep(100 * time.Millisecond)
		i >>= 1
	}
	a.Close()
}
