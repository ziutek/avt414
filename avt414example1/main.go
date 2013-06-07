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
	defer a.Close()

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
		if i == 0 {
			i = 0x80
		}
		start := time.Now()
		for n := 0; n < loop; n++ {
			if err = a.Write(port, i); err != nil {
				log.Fatal(err)
			}
			if err = a.Write(port, 0); err != nil {
				log.Fatal(err)
			}
		}
		wr := time.Now().Sub(start)
		start = time.Now()
		var b byte
		for n := 0; n < loop; n++ {
			b, err = a.Read('B')
			if err != nil {
				log.Fatal(err)
			}
		}
		rd := time.Now().Sub(start)
		start = time.Now()
		var c uint16
		for n := 0; n < loop; n++ {
			c, err = a.ADC(0)
			if err != nil {
				log.Fatal(err)
			}
		}
		ad := time.Now().Sub(start)
		log.Printf(
			"(wr/s=%d), %x (rd/s=%d), %.3f (ad/s=%d)",
			2*loop*time.Second/wr,
			b, loop*time.Second/rd,
			2.5*float32(c)/1023, loop*time.Second/ad,
		)
		i >>= 1
	}
}
