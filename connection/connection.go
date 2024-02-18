package connection

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	def "unolink-client/definitions"

	"github.com/go-resty/resty/v2"
)

const (
	LIST_MAPPING_REFRESH      = 1 * time.Second
	TELEMETRY_MAPPING_REFRESH = 1 * time.Second
)

var (
	address    string
	streamPort uint16
	restPort   uint16
)

func Handle(ctx context.Context, wg *sync.WaitGroup, errorCh chan<- error, quitCh <-chan struct{}, addr string, streamP, restP uint16) {
	address = addr
	restPort = restP
	streamPort = streamP

	defer wg.Done()
	wg.Add(3)

	go handleList(ctx, wg, errorCh, quitCh)
	go handleMapping(ctx, wg, errorCh, quitCh)
	go handleStream(ctx, wg, errorCh, quitCh)
}

// https://www.sobyte.net/post/2021-06/tutorials-for-using-resty/
func handleList(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, quitCh <-chan struct{}) {
	defer wg.Done()
	url := fmt.Sprintf("http://"+address+":%d", restPort)
	client := resty.New()

	type ListDevicesResp struct {
		Result string            `json:"result"`
		Infos  []def.ListDevices `json:"infos"`
	}

	req := client.R()
	for {
		select {
		case <-quitCh:
			// fmt.Println("Quit list")
			return
		case <-ctx.Done():
			return
		default:
			var ListDevicesResp ListDevicesResp
			resp, err := req.Get(url + "/listDevices")
			if err == nil {
				err := json.Unmarshal(resp.Body(), &ListDevicesResp)
				if err == nil {
					def.List = ListDevicesResp.Infos
					def.UpdateDevices()
				}
			} else {
				// fmt.Println("Error      :", err)
				errCh <- err
				return
			}
			time.Sleep(LIST_MAPPING_REFRESH)
		}
	}
}

func handleMapping(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, quitCh <-chan struct{}) {
	defer wg.Done()
	url := fmt.Sprintf("http://"+address+":%d", restPort)
	client := resty.New()

	type TelemetryMappingResp struct {
		Result  string           `json:"result"`
		Mapping map[string]uint8 `json:"mapping"`
	}

	req := client.R()
	for {
		select {
		case <-quitCh:
			// fmt.Println("Quit mapping")
			return
		case <-ctx.Done():
			return
		default:
			var TelemetryMappingResp TelemetryMappingResp
			resp, err := req.Get(url + "/getTelemetryMapping")
			if err == nil {
				err := json.Unmarshal(resp.Body(), &TelemetryMappingResp)
				if err == nil {
					def.TelemetryMapping = TelemetryMappingResp.Mapping
				}
			} else {
				errCh <- err
				return
			}
			time.Sleep(TELEMETRY_MAPPING_REFRESH)
		}
	}
}

func handleStream(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error, quitCh <-chan struct{}) {
	defer wg.Done()
	tcpAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(address+":%d", streamPort))
	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		errCh <- err
		return
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	bufio.NewReader(conn)

	// handle incoming packets
	for {
		select {
		case <-quitCh:
			// fmt.Println("Quit stream")
			return
		case <-ctx.Done():
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			reply := make([]byte, 22)
			_, err = conn.Read(reply)
			// if err != nil {
			// 	errCh <- err
			// 	return
			// }
			if err == nil {
				def.DecodePacket(reply)
			}
		}
	}
}

func Activate(devices []string) {
	var url = fmt.Sprintf("http://"+address+":%d/activate?devices=", restPort)
	for _, device := range devices {
		url += device + "+"
	}
	url = url[:len(url)-1] // remove last '+'
	client := resty.New()
	_, err := client.R().Get(url)
	if err != nil {
		fmt.Println("Error during activation: ", err)
	}
}

func Deactivate(devices []string) {
	var url = fmt.Sprintf("http://"+address+":%d/deactivate?devices=", restPort)
	for _, device := range devices {
		url += device + "+"
	}
	url = url[:len(url)-1] // remove last '+'
	client := resty.New()
	_, err := client.R().Get(url)
	if err != nil {
		fmt.Println("Error during deactivation: ", err)
	}
}

func Shutdown(devices []string) {
	var url = fmt.Sprintf("http://"+address+":%d/shutdown?devices=", restPort)
	for _, device := range devices {
		url += device + "+"
	}
	url = url[:len(url)-1] // remove last '+'
	client := resty.New()
	_, err := client.R().Get(url)
	if err != nil {
		fmt.Println("Error during shutdown: ", err)
	}
}

func ToggleTelemetry(device string) {
	var url string
	_, ok := def.TelemetryMapping[device]
	if ok {
		url = fmt.Sprintf("http://"+address+":%d/exitTelemetry?devices=%s", restPort, device)
	} else {
		url = fmt.Sprintf("http://"+address+":%d/startTelemetry?devices=%s&VO2Max=18.18", restPort, device)
	}

	client := resty.New()
	_, err := client.R().Get(url)
	if err != nil {
		fmt.Println("Error toggling telemetry: ", err)
	}
}

func StartTelemetry(devices []string) {
	var url = fmt.Sprintf("http://"+address+":%d/startTelemetry?devices=", restPort)
	for _, device := range devices {
        url += device + "+"
	}
	url = url[:len(url)-1] // remove last '+'
    url += "&"

    for range devices {
        url += "VO2Max=18.18+"
    }
	url = url[:len(url)-1] // remove last '+'

    fmt.Println(url)

	client := resty.New()
	_, err := client.R().Get(url)
	if err != nil {
		fmt.Println("Error starting telemetry", err)
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
