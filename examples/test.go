package main

import (
	"github.com/ziutek/avt414"
	"log"
	"time"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	a, err := avt414.Open("/dev/ttyUSB0")
	checkErr(err)

	port := byte('D')
	checkErr(a.Setup(port, 0))

	buf := []byte("blaaaaa")
	n, err := a.Writer(port).Write(buf)
	checkErr(err)
	if n != len(buf) {
		panic("bad n")
	}

	const loop = 20
	var i byte
	for {
		start := time.Now()
		if i == 0 {
			i = 0x80
		}
		for n := 0; n < loop; n++ {
			if err = a.Write(port, i); err != nil {
				log.Fatal(err)
			}
		}
		var b byte
		for n := 0; n < loop; n++ {
			b, err = a.Read('B')
			if err != nil {
				log.Fatal(err)
			}
		}
		var c uint16
		for n := 0; n < loop; n++ {
			c, err = a.ADC(0)
			if err != nil {
				log.Fatal(err)
			}
		}
		dt := time.Now().Sub(start)
		log.Printf(
			"%x, %.3f  %d cmd/s",
			b, 2.5*float32(c)/1023, 3*loop*time.Second/dt,
		)
		i >>= 1
	}
	a.Close()
}
