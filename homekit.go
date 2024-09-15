package rfplayer

import (
	"github.com/brutella/hap/accessory"
)

// RFDeviceAccessory represents a HomeKit accessory for an RF device.
type RFDeviceAccessory struct {
	*accessory.A
	Switch   *accessory.Switch
	rf       *RFPlayer
	id       int
	protocol string
}

// NewRFDeviceAccessory creates a new RFDeviceAccessory instance.
func NewRFDeviceAccessory(info accessory.Info, rf *RFPlayer, id int, protocol string) (*RFDeviceAccessory, error) {
	a := RFDeviceAccessory{
		A:        accessory.New(info, accessory.TypeSwitch),
		Switch:   accessory.NewSwitch(info),
		rf:       rf,
		id:       id,
		protocol: protocol,
	}

	var gerr error
	a.Switch.Switch.On.OnValueRemoteUpdate(func(on bool) {
		action := "OFF"
		if on {
			action = "ON"
		}
		_, err := a.rf.EmitSignal(a.protocol, a.id, action)
		if err != nil {
			gerr = err
		}
	})

	return &a, gerr
}
