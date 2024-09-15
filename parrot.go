package rfplayer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParrotDevice represents a Parrot device from the RFPlayer.
type ParrotDevice struct {
	ID       int
	Name     string
	Protocol string
}

// GetParrotDevices returns a list of Parrot devices from the RFPlayer.
func GetParrotDevices(rf *RFPlayer) ([]ParrotDevice, error) {
	status, err := rf.GetStatus("PARROT", "JSON")
	if err != nil {
		return nil, fmt.Errorf("failed to get Parrot status: %v", err)
	}

	var statusData map[string]interface{}
	if err = json.Unmarshal([]byte(status), &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse Parrot status: %v", err)
	}

	var devices []ParrotDevice

	parrotStatus, ok := statusData["parrotStatus"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid Parrot status format")
	}

	for key, value := range parrotStatus {
		if strings.HasPrefix(key, "entry") {
			entry, ok := value.(map[string]interface{})
			if !ok {
				continue
			}

			id, _ := strconv.Atoi(strings.TrimPrefix(key, "entry"))
			name, _ := entry["reminder"].(string)
			protocol, _ := entry["protocol"].(string)

			devices = append(devices, ParrotDevice{
				ID:       id,
				Name:     name,
				Protocol: protocol,
			})
		}
	}

	return devices, nil
}
