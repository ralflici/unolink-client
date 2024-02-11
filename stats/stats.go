package stats

import (
	"encoding/binary"
	def "unolink-client/definitions"
)

func CreateDevice(addr [3]uint8) *def.DeviceState {
	device := def.DeviceState{
		Id: addr,
        Slot: 0,
        LiveOn: false,
		Counter: def.PacketCounter{
			NumInstantaneous: 0,
			NumCumulative:    0,
			NumPosition:      0,
			NumOtherData1:    0,
			NumOtherData2:    0,
			NumOtherData3:    0,
		},
		Time:          0,
		Speed:         0,
		Hrm:           0,
		Power:         0.0,
		Vo2:           0.0,
		Energy:        0.0,
		Distance:      0.0,
		EquivDistance: 0.0,
		PeCounter:     0,
		Acc:           0,
		Dec:           0,
		Jump:          0,
		Impact:        0,
		CumDistance:   [5]uint32{0, 0, 0, 0, 0},
		Hmld:          0,
		TagId:         0,
		Lat:           0,
		Lng:           0,
	}
	def.Devices = append(def.Devices, device)
	return &device
}

func GetDevice(addr def.RadioAddress) *def.DeviceState {
	for i := range def.Devices {
		if def.Devices[i].AddressMatches(addr) {
			return &def.Devices[i]
		}
	}
	return nil
}

func updatePacket(last *uint32, current []uint8) {
	*last = binary.LittleEndian.Uint32(current)
}

func DecodePacket(packet []byte) def.DeviceState {
	deviceID := def.RadioAddress{packet[3], packet[2], packet[1]}
	device := GetDevice(deviceID)
	if device == nil {
		device = CreateDevice(deviceID)
	}

	switch packet[0] {
	case def.Cumulative:
        device.UpdateCumulative(packet[:])
	case def.Instantaneous:
        device.UpdateInstantaneous(packet[:])
	case def.Position:
        device.UpdatePosition(packet[:])
	case def.OtherData1:
        device.UpdateOtherData1(packet[:])
	case def.OtherData2:
        device.UpdateOtherData2(packet[:])
	case def.OtherData3:
        device.UpdateOtherData3(packet[:])
	}
	return *device
}
