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
    go handleREST(address, restPort)
    go handleStream(address, streamPort)
}

func handleREST(address string, restPort uint16) {
    // url := fmt.Sprintf(address+":%d", restPort)
    // Create a Resty Client
    client := resty.New()

    req := client.R()

    // https://www.sobyte.net/post/2021-06/tutorials-for-using-resty/
    go func() {
        for {
            resp, err := req.Get("https://jsonplaceholder.typicode.com/todos/4")

            // Explore response object
            fmt.Println("Response Info:")
            fmt.Println("  Error      :", err)
            fmt.Println("  Status Code:", resp.StatusCode())
            fmt.Println("  Status     :", resp.Status())
            fmt.Println("  Proto      :", resp.Proto())
            fmt.Println("  Time       :", resp.Time())
            fmt.Println("  Received At:", resp.ReceivedAt())
            fmt.Println("  Body       :\n", resp)
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
