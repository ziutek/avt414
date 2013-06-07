package avt414

import (
	"fmt"
	"github.com/ziutek/serial"
	"io"
)

const esc = 0x1b

var hex = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'A', 'B', 'C', 'D', 'E', 'F'}

type AVT414 struct {
	s     *serial.Serial
	lastR bool // last operation was read from binary port
	lastW byte // last written port
	lastA byte // last readed ADC channel
	cmd   [5]byte
}

// Open opens an AVT414 device.
// name - path to the device file (eg: /dev/ttyUSB0)
func Open(name string) (*AVT414, error) {
	s, err := serial.Open(name)
	if err != nil {
		return nil, err
	}
	if err = s.SetSpeed(57600); err != nil {
		return nil, err
	}
	a := AVT414{s: s}
	a.lastA = 0xff
	a.cmd[0] = esc
	return &a, nil
}

// Close closes a device
func (a *AVT414) Close() error {
	err := a.s.Close()
	a.s = nil
	return err
}

func checkPort(port byte) error {
	if port != 'B' && port != 'C' && port != 'D' {
		return fmt.Errorf("Unknown AVT414 port: %c", port)
	}
	return nil
}

// Setup setups lines in specified port for input/output
//  port:  'B', 'C' or 'D'
//  iomask: 0 - output, 1 - input
func (a *AVT414) Setup(port, iomask byte) error {
	err := checkPort(port)
	if err != nil {
		return err
	}
	if port == 'C' {
		// Protect C6 and C7 because they are used to communicate with CPU
		iomask &^= 0x40
		iomask |= 0x80
	}
	// Reset internal state
	a.lastR = false
	a.lastW = 0
	a.lastA = 0xff
	// Write a setup command
	a.cmd[1] = 'S'
	a.cmd[2] = port
	a.cmd[3] = hex[(iomask>>4)&0xf]
	a.cmd[4] = hex[iomask&0xf]
	_, err = a.s.Write(a.cmd[:5])
	return err
}

// Write writes one byte to the port. It uses two byte abbreviated (fast)
// command for subsequent writes to the same port. See Writer for write byte
// slice to the same port.
func (a *AVT414) Write(port, val byte) error {
	err := checkPort(port)
	if err != nil {
		return err
	}
	// Common parameters for two kinds of wirite command
	a.cmd[3] = hex[(val>>4)&0xf]
	a.cmd[4] = hex[val&0xf]
	if a.lastW == port {
		// Write an abbreviated command
		_, err = a.s.Write(a.cmd[3:5])
	} else {
		a.lastR = false
		a.lastW = port
		a.lastA = 0xff
		// Write a full command
		a.cmd[1] = 'W'
		a.cmd[2] = port
		_, err = a.s.Write(a.cmd[:5])
	}
	return err
}

// Read reads from port. It uses one byte abbreviated command for subsequent
// reads.
func (a *AVT414) Read(port byte) (byte, error) {
	err := checkPort(port)
	if err != nil {
		return 0, err
	}
	// Common parameter for two kinds of read command
	a.cmd[2] = port
	if a.lastR {
		// Write an abbreviated command
		_, err = a.s.Write(a.cmd[2:3])
	} else {
		a.lastR = true
		a.lastW = 0
		a.lastA = 0xff
		// Write a full command
		a.cmd[1] = 'R'
		a.cmd[3] = 'B'
		_, err = a.s.Write(a.cmd[:4])
	}
	if err != nil {
		return 0, err
	}
	return a.s.ReadByte()
}

// Reads 10-bit value from specified ADC line (0 <= line <= 7). It uses one
// byte abbreviated command for subsequent reads.
func (a *AVT414) ADC(line byte) (uint16, error) {
	if line > 7 {
		return 0, fmt.Errorf("Bad AVT414 ADC line number: %d", line)
	}
	var err error
	// Common parameter for two kinds of ADC command
	a.cmd[3] = hex[line]
	if a.lastA == line {
		// Write an abbreviated command
		_, err = a.s.Write(a.cmd[3:4])
	} else {
		a.lastR = false
		a.lastW = 0
		a.lastA = line
		// Write a full command
		a.cmd[1] = 'A'
		a.cmd[2] = 'B'
		_, err = a.s.Write(a.cmd[:4])
	}
	if err != nil {
		return 0, err
	}
	n, err := a.s.Read(a.cmd[3:5])
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_, err = a.s.Read(a.cmd[4:5])
		if err != nil {
			return 0, err
		}
	}
	return (uint16(a.cmd[3]) << 8) | uint16(a.cmd[4]), nil
}

type writer struct {
	port byte
	a    *AVT414
}

func (w writer) Write(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return
	}
	err = w.a.Write(w.port, buf[0])
	if err != nil {
		return
	}
	n++
	var (
		b   byte
		cmd = w.a.cmd[3:5]
		s   = w.a.s
	)
	for _, b = range buf[1:] {
		cmd[0] = hex[(b>>4)&0xf]
		cmd[1] = hex[b&0xf]
		_, err = s.Write(cmd)
		if err != nil {
			return
		}
		n++
	}
	return
}

// Writer returns io.Writer that can be used for fast write byte slices to the
// specified port.
func (a *AVT414) Writer(port byte) io.Writer {
	return writer{port, a}
}
