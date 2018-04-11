package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

var debug *bool
var count *uint
var write bool
var address uint64

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [-d] [-c X] <address> [write value] .. [write value]\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
	}
	debug = flag.Bool("d", false, "debug info")
	count = flag.Uint("c", 1, "read count")
	flag.Parse()
}

func main() {
	switch flag.NArg() {
	case 0:
		flag.Usage()
		return
	case 1:
		address, _ = strconv.ParseUint(flag.Arg(0), 0, 16)
	default:
		write = true
		address, _ = strconv.ParseUint(flag.Arg(0), 0, 16)
	}

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	// Use spireg SPI port registry to find the first available SPI bus.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	c, err := p.Connect(32000000, spi.Mode0, 8)
	if err != nil {
		log.Fatal(err)
	}

	// Prints out the gpio pin used.
	if *debug {
		if p, ok := c.(spi.Pins); ok {
			fmt.Printf("  CLK : %s\n", p.CLK())
			fmt.Printf("  MOSI: %s\n", p.MOSI())
			fmt.Printf("  MISO: %s\n", p.MISO())
			fmt.Printf("  CS  : %s\n", p.CS())
		}
	}

	// read / write
	var i uint64
	i = 1 // inc
	var send []byte
	var x uint
	if write {
		send = make([]byte, 2+((flag.NArg()-1)*2))
		send[0] = byte((address >> 7) & 0x0FF)
		send[1] = byte(((address << 1) & 0xFC) | (i << 1))
		for x = 0; x < (uint(flag.NArg()) - 1); x++ {
			y, _ := strconv.ParseUint(flag.Arg(int(x)+1), 0, 16)
			send[2+x*2] = byte(y)
			send[3+x*2] = byte(y >> 8)
			fmt.Printf("Writing 0x%04x @ 0x%04x\n", binary.LittleEndian.Uint16(send[2+x*2:]), address+uint64(x*2))
		}
	} else {
		send = make([]byte, 2+(*count*2))
		send[0] = byte((address >> 7) & 0x0FF)
		send[1] = byte(((address << 1) & 0xFC) | 0x01 | (i << 1))
	}
	if *debug {
		fmt.Printf("Before: %#v\n", send)
	}
	if err := c.Tx(send, send); err != nil {
		log.Fatal(err)
	}
	if *debug {
		fmt.Printf("After: %#v\n", send)
	}
	if !write {
		for x = 0; x < *count; x++ {
			fmt.Printf("Reading 0x%04x @ 0x%04x\n", binary.LittleEndian.Uint16(send[2+x*2:]), address+uint64(x*2))
		}
	}
}
