package connection

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"unolink-client/stats"

	"github.com/go-resty/resty/v2"
)

func Handle(address string, streamPort, restPort uint16) {
    // go handleREST(address, restPort)
    go handleStream(address, streamPort)
}

// https://www.sobyte.net/post/2021-06/tutorials-for-using-resty/
func handleREST(address string, restPort uint16) {
    url := fmt.Sprintf("http://"+address+":%d", restPort)
    // Create a Resty Client
    client := resty.New()


    go func() {
        req := client.R()
        for {
            resp, err := req.Get(url + "/listDevices")
            if err != nil {
                fmt.Println("  Error      :", err)
            }
            fmt.Println("  Body       :\n", resp.String())
            fmt.Println()

            time.Sleep(3 * time.Second)
        }
    }()

    go func() {
        req := client.R()
        for {
            resp, err := req.Get(url + "/getTelemetryMapping")
            if err != nil {
                fmt.Println("  Error      :", err)
            }
            fmt.Println("  Body       :\n", resp.String())
            fmt.Println()

            time.Sleep(3 * time.Second)
        }
    }()

    select {}
}

func handleStream(address string, streamPort uint16) {
    tcpAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(address+":%d", streamPort))
    conn, err := net.DialTCP("tcp", nil, tcpAddress)
    if err != nil {
        fmt.Println("Error connecting to server", err)
        os.Exit(1)
        return
    }
    fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
    bufio.NewReader(conn)

    // handle incoming packets
    for {
        reply := make([]byte, 22)
        _, err = conn.Read(reply)
        if err != nil {
            fmt.Println("Error reading from connection", err)
            // os.Exit(1)
            return
        }
        stats.DecodePacket(reply)
    }
}
