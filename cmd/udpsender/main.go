package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	ln, err := net.ResolveUDPAddr("udp", ":42069")

	if err != nil {
		panic(err)
	}

	udpConn, err := net.DialUDP("udp", nil, ln)
	if err != nil {
		panic(err)
	}

	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")

		str, err := reader.ReadBytes('\n')

		if err != nil && err != io.EOF {
			panic(err)
		}

		udpConn.Write(str)

	}

}
