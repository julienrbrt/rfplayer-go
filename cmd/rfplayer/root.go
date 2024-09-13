package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	rfplayer "github.com/julienrbrt/rfplayer-go"

	"github.com/spf13/cobra"
)

// RootCmd returns the root command for the rfplayer CLI.
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rfplayer",
		Short: "RFPlayer CLI",
		Long:  `A command line interface for interacting with the RFPlayer device.`,
	}

	rootCmd.PersistentFlags().StringVar(&port, "port", "/dev/ttyUSB0", "Serial port for RFPlayer")
	rootCmd.AddCommand(
		helloCmd(),
		pingCmd(),
		emitCmd(),
		recordCmd(),
		listenCmd(),
		setFreqCmd(),
		statusCmd(),
	)

	return rootCmd
}

var (
	port     string
	protocol string
	id       int
	action   string
	band     string
	freq     int
	metadata string
)

func initRFPlayer() (*rfplayer.RFPlayer, error) {
	rf, err := rfplayer.New(port)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RFPlayer: %v", err)
	}
	return rf, nil
}

func helloCmd() *cobra.Command {
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Send HELLO command to RFPlayer",
		Long:  "Send HELLO command to RFPlayer to get device information",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			response, err := rf.Hello()
			if err != nil {
				return fmt.Errorf("HELLO command failed: %v", err)
			}

			cmd.Println(response)

			return nil
		},
	}

	return helloCmd
}

func pingCmd() *cobra.Command {
	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping the RFPlayer device",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			start := time.Now()
			err = rf.Ping()
			duration := time.Since(start)

			if err != nil {
				return fmt.Errorf("ping failed: %v", err)
			}

			cmd.Printf("Ping successful! Response time: %v\n", duration)
			return nil
		},
	}

	return pingCmd
}

func emitCmd() *cobra.Command {
	emitCmd := &cobra.Command{
		Use:   "emit",
		Short: "Emit a signal",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			resp, err := rf.EmitSignal(protocol, id, action)
			if err != nil {
				return fmt.Errorf("failed to emit signal: %v", err)
			}
			cmd.Println(resp)

			cmd.Println("Signal emitted successfully")

			return nil
		},
	}
	emitCmd.Flags().StringVar(&protocol, "protocol", "", "Protocol to use")
	emitCmd.Flags().IntVar(&id, "id", 0, "ID to use")
	emitCmd.Flags().StringVar(&action, "action", "ON", "Action to perform (ON/OFF)")

	return emitCmd
}

func recordCmd() *cobra.Command {
	recordCmd := &cobra.Command{
		Use:   "record",
		Short: "Record a signal",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			resp, err := rf.RecordSignal(id, action, metadata)
			if err != nil {
				return fmt.Errorf("failed to record signal: %v", err)
			}

			cmd.Println(resp)
			cmd.Println("Signal recorded successfully")

			return nil
		},
	}

	recordCmd.Flags().IntVar(&id, "id", 0, "ID to use")
	recordCmd.Flags().StringVar(&action, "action", "ON", "Action to record (ON/OFF)")
	recordCmd.Flags().StringVar(&metadata, "metadata", "", "Metadata to record")

	return recordCmd
}

func listenCmd() *cobra.Command {
	listenCmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for signals",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			cmd.Println("Listening for signals... Press Ctrl+C to stop.")

			if err := rf.StartListening(func(signal string) {
				cmd.Println("Received signal:", signal)
			}); err != nil {
				return fmt.Errorf("error while listening: %v", err)
			}

			return nil
		},
	}
	return listenCmd
}

func setFreqCmd() *cobra.Command {
	setFreqCmd := &cobra.Command{
		Use:   "setfreq",
		Short: "Set frequency",
		Long: `Set frequency for the RFPlayer. The frequency is specified in KHz.
The available bands are L (0 (disabled), 433420 and 433920) and H (0 (disabled), 868950 and 868350)`,
		Example: `rfplayer setfreq --band L --freq 433920`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			if freq < 0 {
				return fmt.Errorf("frequency must be positive")
			}

			resp, err := rf.SetFrequency(band, freq)
			if err != nil {
				return fmt.Errorf("failed to set frequency: %v", err)
			}
			cmd.Println(resp)

			return nil
		},
	}

	setFreqCmd.Flags().StringVar(&band, "band", "L", "Frequency band (L/H)")
	setFreqCmd.Flags().IntVar(&freq, "freq", 433920, "Frequency in KHz")

	return setFreqCmd
}

func statusCmd() *cobra.Command {
	var statusType string
	var format string

	statusCmd := &cobra.Command{
		Use:   "status [type]",
		Short: "Get RFPlayer status",
		Long: `Get RFPlayer status. Available types:
- SYSTEM: General system information
- RADIO: Radio-related information
- TRANSCODER: Transcoder configuration
- PARROT: Parrot configuration
- ALARM: Alarm configuration
You can also specify the output format: TEXT (default), XML, or JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			status, err := rf.GetStatus(statusType, format)
			if err != nil {
				return fmt.Errorf("failed to get status: %v", err)
			}

			// improve output formatting
			if strings.EqualFold(format, "JSON") {
				// pretty print JSON
				data := map[string]interface{}{}
				if err := json.Unmarshal([]byte(status), &data); err != nil {
					return fmt.Errorf("failed to unmarshal JSON: %v", err)
				}

				statusBz, err := json.Marshal(data)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}

				status = string(statusBz)
			}

			cmd.Println(status)

			return nil
		},
	}

	statusCmd.Flags().StringVar(&statusType, "type", "", "Type of status to retrieve (SYSTEM, RADIO, TRANSCODER, PARROT, ALARM)")
	statusCmd.Flags().StringVar(&format, "format", "TEXT", "Output format (TEXT, XML, JSON)")

	return statusCmd
}
