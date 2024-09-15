package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
	rfplayer "github.com/julienrbrt/rfplayer-go"
	"github.com/spf13/cobra"
)

// global flags
var (
	port   string
	id     int
	action string
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
		setFreqCmd(),
		statusCmd(),
		homekitCmd(),
		factoryResetCmd(),
	)

	return rootCmd
}

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
		Short: "Display device information",
		Long:  "Send HELLO command to RFPlayer to get device information",
		Args:  cobra.NoArgs,
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
		Args:  cobra.NoArgs,
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
	var protocol string

	emitCmd := &cobra.Command{
		Use:   "emit",
		Short: "Emit a signal",
		Long:  `Emit a signal using the RFPlayer.`,
		Args:  cobra.NoArgs,
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
	var metadata string

	recordCmd := &cobra.Command{
		Use:   "record",
		Short: "Record a signal using Parrot",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			_, err = rf.RecordSignal(id, action, metadata)
			if err != nil {
				return fmt.Errorf("failed to record signal: %v", err)
			}
			cmd.Println("Signal is being recorded... Press physical button on remote control to cancel capture.")

			return nil
		},
	}

	recordCmd.Flags().IntVar(&id, "id", 0, "ID to use")
	recordCmd.Flags().StringVar(&action, "action", "ON", "Action to record (ON/OFF)")
	recordCmd.Flags().StringVar(&metadata, "metadata", "", "Metadata to record to help identify the signal later")

	return recordCmd
}

func setFreqCmd() *cobra.Command {
	var (
		band string
		freq int
	)

	setFreqCmd := &cobra.Command{
		Use:   "setfreq",
		Short: "Set frequency for the RFPlayer based on the band",
		Long: `Set frequency for the RFPlayer. The frequency is specified in KHz.
The available bands are L (0 (disabled), 433420 and 433920) and H (0 (disabled), 868950 and 868350)`,
		Example: `rfplayer setfreq --band L --freq 433920`,
		Args:    cobra.NoArgs,
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
	setFreqCmd.MarkFlagRequired("band")
	setFreqCmd.MarkFlagRequired("freq")

	return setFreqCmd
}

func statusCmd() *cobra.Command {
	var format string

	statusCmd := &cobra.Command{
		Use:   "status [type]",
		Short: "Display RFPlayer status",
		Long: `Get RFPlayer status. Available types:
- SYSTEM: General system information
- RADIO: Radio-related information
- TRANSCODER: Transcoder configuration
- PARROT: Parrot configuration
- ALARM: Alarm configuration
You can also specify the output format: TEXT (default), XML, or JSON`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			statusType := ""
			if len(args) > 0 {
				statusType = args[0]
			}

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

	statusCmd.Flags().StringVar(&format, "format", "TEXT", "Output format (TEXT, XML, JSON)")

	return statusCmd
}

func factoryResetCmd() *cobra.Command {
	var all bool

	factoryResetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Perform a factory reset on the RFPlayer",
		Long: `Perform a factory reset on the RFPlayer. 
Use the --all flag to reset everything including PARROT records and TRANSCODER configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			// Prompt for confirmation
			cmd.Print("Are you sure you want to perform a factory reset? This action cannot be undone. (y/N): ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) != "y" {
				cmd.Println("Factory reset cancelled.")
				return nil
			}

			cmd.Println("Performing factory reset...")
			err = rf.FactoryReset(all)
			if err != nil {
				return fmt.Errorf("factory reset failed: %v", err)
			}

			if all {
				cmd.Println("Full factory reset completed successfully. All settings, including PARROT records and TRANSCODER configuration, have been reset.")
			} else {
				cmd.Println("Factory reset completed successfully. Note that PARROT records and TRANSCODER configuration were not affected.")
			}

			return nil
		},
	}

	factoryResetCmd.Flags().BoolVar(&all, "all", false, "Reset everything, including PARROT records and TRANSCODER configuration")

	return factoryResetCmd
}

func homekitCmd() *cobra.Command {
	var pin string

	homekitCmd := &cobra.Command{
		Use:   "homekit",
		Short: "Start HomeKit server",
		RunE: func(cmd *cobra.Command, args []string) error {
			rf, err := initRFPlayer()
			if err != nil {
				return err
			}
			defer rf.Close()

			devices, err := rfplayer.GetParrotDevices(rf)
			if err != nil {
				return fmt.Errorf("failed to get Parrot devices: %v", err)
			}

			var accessories []*accessory.A
			for _, dev := range devices {
				info := accessory.Info{
					Name:         dev.Name,
					SerialNumber: fmt.Sprintf("%d", dev.ID),
					Manufacturer: "RFPlayer",
					Model:        dev.Protocol,
				}
				acc, err := rfplayer.NewRFDeviceAccessory(info, rf, dev.ID, dev.Protocol)
				if err != nil {
					return fmt.Errorf("failed to create RF device accessory: %v", err)
				}

				accessories = append(accessories, acc.A)
			}

			if len(accessories) == 0 {
				return fmt.Errorf("no devices found in Parrot memory")
			}

			bridge := accessory.New(accessory.Info{
				Name:         "RFPlayer Bridge",
				SerialNumber: "RF0001",
				Manufacturer: "RFPlayer",
				Model:        "Bridge",
			}, accessory.TypeBridge)

			fs := hap.NewFsStore("./db")
			server, err := hap.NewServer(fs, bridge, accessories...)
			if err != nil {
				return fmt.Errorf("failed to create HomeKit server: %v", err)
			}

			server.Pin = pin

			cmd.Printf("HomeKit server is running with %d devices, PIN: %s", len(accessories), pin)
			server.ListenAndServe(cmd.Context())

			return nil
		},
	}

	homekitCmd.Flags().StringVar(&pin, "pin", "00102003", "HomeKit PIN")

	return homekitCmd
}
