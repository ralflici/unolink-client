package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	def "unolink-client/definitions"

	"github.com/go-resty/resty/v2"
)

const (
	LIST_MAPPING_REFRESH      = 10 * time.Second
	TELEMETRY_MAPPING_REFRESH = 3 * time.Second
)

var (
    address string
    streamPort uint16
    restPort uint16
)

func Handle(addr string, streamP, restP uint16) {
    address = addr
    restPort = restP
    streamPort = streamP

	def.WaitGroup.Add(1)
	go handleREST()

	def.WaitGroup.Add(1)
	go handleStream()

	def.WaitGroup.Wait()
}

// https://www.sobyte.net/post/2021-06/tutorials-for-using-resty/
func handleREST() {
	url := fmt.Sprintf("http://"+address+":%d", restPort)
	client := resty.New()

	def.WaitGroup.Add(1)
	go func() {
		defer def.WaitGroup.Done()
		type ListDevicesResp struct {
			Result string            `json:"result"`
			Infos  []def.ListDevices `json:"infos"`
		}

		req := client.R()
		for {
			var ListDevicesResp ListDevicesResp
			resp, err := req.Get(url + "/listDevices")
			if err == nil {
				err := json.Unmarshal(resp.Body(), &ListDevicesResp)
				if err == nil {
					def.List = ListDevicesResp.Infos
                    def.UpdateDevices()
					// fmt.Println("List       :", def.List)
				} else {
					// fmt.Println("Error parsing JSON", err)
				}
			} else {
				fmt.Println("Error      :", err)
				def.WaitGroup.Done()
				return
			}

			time.Sleep(LIST_MAPPING_REFRESH)
		}
	}()

	def.WaitGroup.Add(1)
	go func() {
		defer def.WaitGroup.Done()
		type TelemetryMappingResp struct {
			Result  string           `json:"result"`
			Mapping map[string]uint8 `json:"mapping"`
		}

		req := client.R()
		for {
			var TelemetryMappingResp TelemetryMappingResp
			resp, err := req.Get(url + "/getTelemetryMapping")
			if err == nil {
				err := json.Unmarshal(resp.Body(), &TelemetryMappingResp)
				if err == nil {
					def.TelemetryMapping = TelemetryMappingResp.Mapping
                    // fmt.Println("Mapping    :", def.TelemetryMapping)
				} else {
					// fmt.Println("Error parsing JSON", err)
				}
			} else {
				fmt.Println("Error      :", err)
				def.WaitGroup.Done()
				return
			}

			time.Sleep(TELEMETRY_MAPPING_REFRESH)
		}
	}()
	def.WaitGroup.Wait()
}

func handleStream() {
	tcpAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(address+":%d", streamPort))
	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		fmt.Println("Error connecting to server", err)
		def.WaitGroup.Done()
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
			def.WaitGroup.Done()
			return
		}
		def.DecodePacket(reply)
	}
}

func Activate(device string) {
	url := fmt.Sprintf("http://"+address+":%d/activate?devices=%s", restPort, device)
	client := resty.New()
    _, err := client.R().Get(url)
    if err != nil {
        fmt.Println("Error activating device %s", err, device)
    } else {
        // fmt.Println("Started telemetry for device:", device, resp)
    }
}

func StartTelemetry(device string) {
    _, ok := def.TelemetryMapping[device]
    if ok {
        fmt.Println("Device already in telemetry")
        return
    }

	url := fmt.Sprintf("http://"+address+":%d/startTelemetry?devices=%s&VO2Max=18.18", restPort, device)
	client := resty.New()
    _, err := client.R().Get(url)
    if err != nil {
        fmt.Println("Error starting telemetry", err)
    } else {
        // fmt.Println("Started telemetry for device:", device, resp)
    }
}

func StopTelemetry() {
	url := fmt.Sprintf("http://"+address+":%d/stopTelemetry", restPort)
	client := resty.New()
    _, err := client.R().Get(url)
    if err != nil {
        fmt.Println("Error stopping telemetry", err)
    } else {
        // fmt.Println("Telemetry stopped")
    }
}

func TelemetryParty() {
	url := fmt.Sprintf("http://"+address+":%d/telemetryParty", restPort)
	client := resty.New()
    _, err := client.R().Get(url)
    if err != nil {
        fmt.Println("Error telemetry party", err)
    } else {
        // fmt.Println("Telemetry stopped")
    }
}

