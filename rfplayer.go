package rfplayer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type RFPlayer struct {
	port *serial.Port
}

// New creates a new RFPlayer instance
func New(portName string) (*RFPlayer, error) {
	config := &serial.Config{
		Name:        portName,
		Baud:        115200,
		ReadTimeout: time.Second * 5,
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port: %v", err)
	}

	return &RFPlayer{port: port}, nil
}

// Close closes the serial connection
func (r *RFPlayer) Close() error {
	return r.port.Close()
}

// Ping sends a PING command to the RFPlayer and checks for a PONG response
func (r *RFPlayer) Ping() error {
	response, err := r.SendCommand("PING")
	if err != nil {
		return fmt.Errorf("failed to send PING command: %v", err)
	}

	if !strings.Contains(response, "PONG") {
		return fmt.Errorf("unexpected response to PING: %s", response)
	}

	return nil
}

// Hello sends the HELLO command to the RFPlayer and returns the response
func (r *RFPlayer) Hello() (string, error) {
	response, err := r.SendCommand("HELLO")
	if err != nil {
		return "", fmt.Errorf("failed to send HELLO command: %v", err)
	}

	return response, nil
}

// SendCommand sends a command to the RFPlayer and returns the response
func (r *RFPlayer) SendCommand(cmd string) (string, error) {
	// Flush input buffer
	r.port.Flush()

	_, err := r.port.Write([]byte("ZIA++" + cmd + "\r"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	reader := bufio.NewReader(r.port)
	response, err := reader.ReadString('\r')
	if err != nil {
		if err == io.EOF {
			return "", nil // EOF is expected when async request is sent
		}

		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Trim the "ZIA--" prefix if present
	response = strings.TrimPrefix(response, "ZIA--")

	return strings.TrimSpace(response), nil
}

// EmitSignal sends a signal using the RFPlayer
func (r *RFPlayer) EmitSignal(protocol string, id int, action string) (string, error) {
	cmd := fmt.Sprintf("%s ID %d %s", action, id, protocol)
	return r.SendCommand(cmd)
}

// RecordSignal puts the RFPlayer in learning mode to record a signal
func (r *RFPlayer) RecordSignal(id int, action, metadata string) (string, error) {
	cmd := fmt.Sprintf("PARROTLEARN ID %d %s [%s]", id, action, metadata)
	return r.SendCommand(cmd)
}

// SetFrequency sets the frequency for the RFPlayer
func (r *RFPlayer) SetFrequency(band string, freq int) (string, error) {
	cmd := fmt.Sprintf("FREQ %s %d", band, freq)
	return r.SendCommand(cmd)
}

// EnableReceiver enables specific protocols for receiving
func (r *RFPlayer) EnableReceiver(protocols ...string) (string, error) {
	cmd := "RECEIVER + " + strings.Join(protocols, " ")
	return r.SendCommand(cmd)
}

// SetFormat sets the format for received RF frames
func (r *RFPlayer) SetFormat(format string) (string, error) {
	cmd := "FORMAT " + format
	return r.SendCommand(cmd)
}

// GetStatus retrieves the status of the RFPlayer
func (r *RFPlayer) GetStatus(statusType string, format string) (string, error) {
	if statusType == "" {
		statusType = "SYSTEM"
	}
	if format == "" {
		format = "TEXT"
	}

	cmd := fmt.Sprintf("STATUS %s %s", statusType, format)
	response, err := r.SendCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get status: %v", err)
	}

	return response, nil
}

// ParrotRemapping remaps Parrot entries to another protocol
func (r *RFPlayer) ParrotRemapping(protocol string, startID int) (string, error) {
	cmd := fmt.Sprintf("REMAPPING PARROT ONOFF %s %s", protocol, idToX10(startID))
	return r.SendCommand(cmd)
}

// FactoryReset performs a factory reset on the RFPlayer
func (r *RFPlayer) FactoryReset(all bool) error {
	var cmd string
	if all {
		cmd = "FACTORYRESET ALL"
	} else {
		cmd = "FACTORYRESET"
	}

	response, err := r.SendCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to send factory reset command: %v", err)
	}

	if !strings.Contains(response, "OK") {
		return fmt.Errorf("unexpected response to factory reset: %s", response)
	}

	return nil
}

// Helper function to convert ID to X10 format
func idToX10(id int) string {
	house := string(rune('A' + (id-1)/16))
	unit := (id-1)%16 + 1
	return fmt.Sprintf("%s%d", house, unit)
}
