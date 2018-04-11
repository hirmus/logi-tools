package main

import (
	"bytes"
	"fmt"
	"os"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

const spiMaxLength = 4096

func spiUpload(buffer []byte, length int) error {
	if _, err := host.Init(); err != nil {
		return err
	}
	pinProg := gpioreg.ByName("GPIO24")
	if pinProg == nil {
		return fmt.Errorf("Failed to find GPIO24")
	}
	pinInit := gpioreg.ByName("GPIO23")
	if pinProg == nil {
		return fmt.Errorf("Failed to find GPIO23")
	}
	if err := pinInit.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	if err := pinProg.Out(gpio.High); err != nil {
		return err
	}
	if err := pinProg.Out(gpio.Low); err != nil {
		return err
	}
	for timer := 0; pinInit.Read() == gpio.High; timer++ {
		if timer > 200 {
			if err := pinProg.Out(gpio.High); err != nil {
				return err
			}
			return fmt.Errorf("FPGA did not answer to prog request, init pin not going low")
		}
	}
	if err := pinProg.Out(gpio.High); err != nil {
		return err
	}
	for timer := 0; pinInit.Read() == gpio.Low; timer++ {
		if timer > 0xffffff {
			if err := pinProg.Out(gpio.High); err != nil {
				return err
			}
			return fmt.Errorf("FPGA did not answer to prog request, init pin not going high")
		}
	}
	p, err := spireg.Open("")
	if err != nil {
		return err
	}
	defer p.Close()
	c, err := p.Connect(8000000, spi.Mode0, 8)
	if err != nil {
		return err
	}
	var send []byte
	var bSize int
	for i := 0; i < length; i += spiMaxLength {
		writeLength := length - i
		if writeLength < spiMaxLength {
			bSize = writeLength
		} else {
			bSize = spiMaxLength
		}
		send = make([]byte, bSize, bSize)
		copy(send, buffer[i:])
		if err := c.Tx(send, send); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage:\n\t%s <.bit file>\n\n", os.Args[0])
		return
	}
	fmt.Println("LOGI_LOADER VERSION : go-0.0.1")
	fmt.Println("for board LOGIPI_R1.0")
	configBits := bytes.Repeat([]byte{0, 0, 0, 0}, 1024*1024)
	fr, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	size, err := fr.Read(configBits)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("bit file size:", size)
	err = spiUpload(configBits, size+5)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println("bitstream loaded, check done led")
}
