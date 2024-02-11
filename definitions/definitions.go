package definitions

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
)

type RadioAddress [3]uint8

func (r RadioAddress) Slice() []byte {
	return r[:]
}

func (r RadioAddress) String() string {
	return fmt.Sprintf("%x%x%x", r[0], r[1], r[2])
}

func RadioAddressFromString(s string) (RadioAddress, error) {
	// Check if the length of the string is a multiple of 2
	if len(s)%2 != 0 {
		return RadioAddress{}, fmt.Errorf("Invalid length")
	}

	// Convert each pair of characters to bytes
	var byteSlice RadioAddress
	for i := 0; i < len(s); i += 2 {
		// Parse the hex value as base 16 and convert it to a byte
		hexValue, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			fmt.Println("Error parsing hexadecimal string:", err)
			return RadioAddress{}, err
		}
		// Append the byte to the slice
		byteSlice[i/2] = byte(hexValue)
	}
	return byteSlice, nil
}

type PacketCounter struct {
	NumInstantaneous uint32
	NumCumulative    uint32
	NumPosition      uint32
	NumOtherData1    uint32
	NumOtherData2    uint32
	NumOtherData3    uint32
}

func (d *PacketCounter) Total() uint32 {
	return d.NumInstantaneous + d.NumCumulative + d.NumPosition +
		d.NumOtherData1 + d.NumOtherData2 + d.NumOtherData3
}

func (d *PacketCounter) Clear() {
	d.NumCumulative = 0
	d.NumInstantaneous = 0
	d.NumPosition = 0
	d.NumOtherData1 = 0
	d.NumOtherData2 = 0
	d.NumOtherData3 = 0
}

func (d *PacketCounter) String() string {
	return fmt.Sprintf("CU = %d, IN = %d, O1 = %d, O2 = %d, O3 = %d, TOT = %d",
		d.NumCumulative, d.NumInstantaneous, d.NumOtherData1,
		d.NumOtherData2, d.NumOtherData3, d.Total())
}

type DeviceState struct {
	Id            RadioAddress
	Slot          uint8
	LiveOn        bool
	Counter       PacketCounter
	Time          uint32
	Speed         uint16
	Hrm           uint8
	Power         float32
	Vo2           float32
	Energy        float32
	Distance      float32
	EquivDistance float32
	PeCounter     uint16
	Acc           uint16
	Dec           uint16
	Jump          uint16
	Impact        uint16
	Hmld          uint32
	CumDistance   [5]uint32
	TagId         uint16
	Lat           uint32
	Lng           uint32
}

func (d *DeviceState) AddressMatches(addr RadioAddress) bool {
	return d.Id == addr
}

func (d *DeviceState) PrintCounter() {
	fmt.Printf("Device %s\tCU = %d, IN = %d, O1 = %d, O2 = %d, O3 = %d, TOT = %d\n",
		d.Id.String(),
		d.Counter.NumCumulative, d.Counter.NumInstantaneous, d.Counter.NumOtherData1,
		d.Counter.NumOtherData2, d.Counter.NumOtherData3, d.Counter.Total())
	d.Counter.Clear()
}

func (d *DeviceState) UpdateCumulative(packet []byte) {
	d.Counter.NumCumulative++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.TagId = binary.LittleEndian.Uint16(packet[7:9])
	d.Energy = math.Float32frombits(binary.LittleEndian.Uint32(packet[9:13]))
	d.Distance = math.Float32frombits(binary.LittleEndian.Uint32(packet[13:17]))
	d.EquivDistance = math.Float32frombits(binary.LittleEndian.Uint32(packet[17:21]))
}

func (d *DeviceState) UpdateInstantaneous(packet []byte) {
	d.Counter.NumInstantaneous++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.Speed = binary.LittleEndian.Uint16(packet[7:9])
	d.Hrm = packet[9]
	d.Power = math.Float32frombits(binary.LittleEndian.Uint32(packet[10:14]))
	d.Vo2 = math.Float32frombits(binary.LittleEndian.Uint32(packet[14:18]))
}

func (d *DeviceState) UpdatePosition(packet []byte) {
	d.Counter.NumPosition++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.Lat = binary.LittleEndian.Uint32(packet[7:11])
	d.Lng = binary.LittleEndian.Uint32(packet[11:15])
}

func (d *DeviceState) UpdateOtherData1(packet []byte) {
	d.Counter.NumOtherData1++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.PeCounter = binary.LittleEndian.Uint16(packet[7:9])
	d.Acc = binary.LittleEndian.Uint16(packet[9:11])
	d.Dec = binary.LittleEndian.Uint16(packet[11:13])
	d.Jump = binary.LittleEndian.Uint16(packet[13:15])
	d.Impact = binary.LittleEndian.Uint16(packet[15:17])
}

func (d *DeviceState) UpdateOtherData2(packet []byte) {
	d.Counter.NumOtherData2++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.CumDistance[0] = binary.LittleEndian.Uint32(append(packet[7:10], []byte{0x00}...))
	d.CumDistance[1] = binary.LittleEndian.Uint32(append(packet[10:13], []byte{0x00}...))
	d.CumDistance[2] = binary.LittleEndian.Uint32(append(packet[13:16], []byte{0x00}...))
	d.CumDistance[3] = binary.LittleEndian.Uint32(append(packet[16:19], []byte{0x00}...))
	d.CumDistance[4] = binary.LittleEndian.Uint32(append(packet[19:22], []byte{0x00}...))
}

func (d *DeviceState) UpdateOtherData3(packet []byte) {
	d.Counter.NumOtherData3++
	d.Time = binary.LittleEndian.Uint32(append(packet[4:7], []byte{0x00}...))
	d.Hmld = binary.LittleEndian.Uint32(append(packet[7:10], []byte{0x00}...))
}

var Devices []DeviceState

const (
	Cumulative    = 0x22
	Instantaneous = 0x23
	Position      = 0x24
	OtherData1    = 0x28
	OtherData2    = 0x2B
	OtherData3    = 0x2C
)
