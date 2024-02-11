package connection

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"unolink-client/stats"
)

func Handle(address string) {
    tcpAddress, err := net.ResolveTCPAddr("tcp", address)
    conn, err := net.DialTCP("tcp", nil, tcpAddress)
    if err != nil {
        fmt.Println("Error connecting to server", err)
        os.Exit(1)
        return
    }
    fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
    bufio.NewReader(conn).ReadString('\n')

    // handle incoming packets
    for {
        reply := make([]byte, 22)
        _, err = conn.Read(reply)
        if err != nil {
            fmt.Println("Error reading from connection", err)
            os.Exit(1)
            return
        }
        stats.DecodePacket(reply)
    }
}
