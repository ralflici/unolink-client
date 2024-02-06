package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/lipgloss/table"
)

type Log struct {
    logType string
    logStyle lipgloss.Style
}

func (l Log) Log(msg string) {
    fmt.Println(l.logStyle.Render(l.logType + msg))
}

type IPAddress      []string
type RadioAddress   [3]uint8
type DeviceStats struct {
    id                  RadioAddress
    num_instantaneous  uint32
    num_cumulative     uint32
    num_position       uint32
    num_otherdata1     uint32
    num_otherdata2     uint32
    num_otherdata3     uint32
}

var (
    Info    Log
    Debug   Log
    Warn    Log
    Error   Log

    devices []DeviceStats
)

const (
    Cumulative = 0x22
    Instantaneous = 0x23
    Position = 0x24
    OtherData1 = 0x28
    OtherData2 = 0x2B
    OtherData3 = 0x2C
)

func initLogs() {
    Info.logType = "INFO: "
    Info.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

    Debug.logType = "DEBUG: "
    Debug.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

    Warn.logType = "WARN: "
    Warn.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

    Error.logType = "ERROR: "
    Error.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
}

func createDevice(addr [3]uint8) *DeviceStats {
    device := DeviceStats{
        id: addr,
        num_instantaneous: 0,
        num_cumulative: 0,
        num_position: 0,
        num_otherdata1: 0,
        num_otherdata2: 0,
        num_otherdata3: 0,
    }
    devices = append(devices, device)
    return &device
}

func getDevice(addr [3]byte) *DeviceStats {
    for i := range devices {
        if devices[i].id == addr {
            return &devices[i]
        }
    }
    return nil
}

func updatePacket(last *uint32, current []uint8) {
    *last = binary.LittleEndian.Uint32(current)
}

func decodePacket(packet []byte) DeviceStats {
    deviceID := RadioAddress{packet[3], packet[2], packet[1]}
    device := getDevice(deviceID)
    if device == nil {
        Debug.Log("Creating device")
        device = createDevice(deviceID)
    }
    // Debug.Log("Found device")

    switch packet[0] {
    case Cumulative:
        device.num_cumulative++
        // updatePacket(&device.last_cumulative, []uint8{packet[4], packet[5], packet[6], 0})
    case Instantaneous:
        device.num_instantaneous++
        // updatePacket(&device.last_instantaneous, []uint8{packet[4], packet[5], packet[6], 0})
    case Position:
        device.num_position++
        // updatePacket(&device.last_position, []uint8{packet[4], packet[5], packet[6], 0})
    case OtherData1:
        device.num_otherdata1++
        // updatePacket(&device.last_otherdata1, []uint8{packet[4], packet[5], packet[6], 0})
    case OtherData2:
        device.num_otherdata2++
        // updatePacket(&device.last_otherdata2, []uint8{packet[4], packet[5], packet[5], 0})
    case OtherData3:
        device.num_otherdata3++
        // updatePacket(&device.last_otherdata3, []uint8{packet[4], packet[5], packet[6], 0})
    }
    return *device
}

func main() {
    initLogs()
    interval := 1 * time.Second

    if len(os.Args) != 2 {
        Error.Log("Usage: go run main.go <address>")
        os.Exit(1)
    }
    address := os.Args[1:][0]
    Debug.Log("Connecting to " + address)

    tcpAddress, err := net.ResolveTCPAddr("tcp", address)
    conn, err := net.DialTCP("tcp", nil, tcpAddress)
    if err != nil {
        Error.Log(err.Error())
        os.Exit(1)
    }
    fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
    bufio.NewReader(conn).ReadString('\n')

    go func() {
        for {
            reply := make([]byte, 22)
            _, err = conn.Read(reply)
            if err != nil {
                Error.Log("Read from server failed: " + err.Error())
                os.Exit(1)
            }
            decodePacket(reply)
        }
    }()

    go func() {
        for {
            fmt.Printf("Number of devices = %d\n", len(devices))
            for i := 0; i < len(devices); i++ {
                fmt.Printf(
                    "Device %x%x%x\tCU = %d, IN = %d, O1 = %d, O2 = %d, O3 = %d, TOT = %d\n",
                    devices[i].id[0], devices[i].id[1], devices[i].id[2],
                    devices[i].num_cumulative,
                    devices[i].num_instantaneous,
                    devices[i].num_otherdata1,
                    devices[i].num_otherdata2,
                    devices[i].num_otherdata3,
                    devices[i].num_cumulative +
                    devices[i].num_instantaneous +
                    devices[i].num_otherdata1 +
                    devices[i].num_otherdata2 +
                    devices[i].num_otherdata3,

                )
                devices[i].num_cumulative = 0;
                devices[i].num_instantaneous = 0;
                devices[i].num_position = 0;
                devices[i].num_otherdata1 = 0;
                devices[i].num_otherdata2 = 0;
                devices[i].num_otherdata3 = 0;
            }
            time.Sleep(interval)
        }
    }()

    select {}
}
