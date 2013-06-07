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

	buf := []byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
	for {
		const loop = 1000
		start := time.Now()
		for i := 0; i < loop; i++ {
			_, err := a.Writer(port).Write(buf)
			checkErr(err)
		}
		dt := time.Now().Sub(start)
		log.Printf("wr/s = %d", time.Duration(len(buf))*loop*time.Second/dt)
	}
}
