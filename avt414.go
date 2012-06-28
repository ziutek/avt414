package avt414

import (
	"fmt"
	"github.com/ziutek/serial"
)

const esc = 0x1b

var hex = [16]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'A', 'B', 'C', 'D', 'E', 'F'}

type Avt414 struct {
	s *serial.Serial
}

// Open connection to the device.
// name - path to the device file (eg: /dev/ttyUSB0)
func Open(name string) (*Avt414, error) {
	s, err := serial.Open(name)
	if err != nil {
		return nil, err
	}
	if err = s.SetSpeed(57600); err != nil {
		return nil, err
	}
	return &Avt414{s}, nil
}

func (a *Avt414) Close() error {
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

// Setups lines in specified port for input/output
//  port: 'B', 'C' or 'D'
//  iomask: 0 - output, 1 - input
func (a *Avt414) Setup(port, iomask byte) error {
	err := checkPort(port)
	if err != nil {
		return err
	}
	if port == 'C' {
		// Protect C6 and C7 because they are used to communicate with CPU
		iomask &^= 0x40
		iomask |= 0x80
	}
	cmd := []byte{
		esc, 'S', port,
		hex[(iomask>>4)&0xf], hex[iomask&0xf],
	}
	_, err = a.s.Write(cmd)
	return err
}

// Write to port
func (a *Avt414) Write(port, val byte) error {
	err := checkPort(port)
	if err != nil {
		return err
	}
	cmd := []byte{
		esc, 'W', port,
		hex[(val>>4)&0xf], hex[val&0xf],
	}
	_, err = a.s.Write(cmd)
	return err
}

// Reads from port
func (a *Avt414) Read(port byte) (byte, error) {
	err := checkPort(port)
	if err != nil {
		return 0, err
	}
	cmd := []byte{esc, 'R', port, 'B'}
	if _, err = a.s.Write(cmd); err != nil {
		return 0, err
	}
	return a.s.ReadByte()
}

// Reads 10-bit value from specified ADC line
//  0 <= line <= 7
func (a *Avt414) ADC(line int) (uint16, error) {
	if line < 0 || line > 7 {
		return 0, fmt.Errorf("Bad AVT414 ADC line number: %d", line)
	}
	cmd := []byte{esc, 'A', 'B', hex[line]}
	_, err := a.s.Write(cmd)
	if err != nil {
		return 0, err
	}
	n, err := a.s.Read(cmd[:2])
	if err != nil {
		return 0, err
	}
	if n == 1 {
		_, err = a.s.Read(cmd[1:2])
		if err != nil {
			return 0, err
		}
	}
	return (uint16(cmd[0]) << 8) | uint16(cmd[1]), nil
}
