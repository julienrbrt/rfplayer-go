# RFPlayer-Go

CLI tool and API for [RFPlayer](https://www.gce-electronics.com/fr/produits-radio/1777-rf-player-3770008041004.html), in Go.
This CLI tool allows to configure the RFPlayer and to send commands to it via the CLI and see the configured devices via HomeKit.

## Installation

```bash
go install github.com/julienrbrt/rfplayer-go/cmd/rfplayer@latest
```

## How to use

Help is available via the command line:

```bash
rfplayer --help
```

Record a signal (_parrot_):

```bash
rfplayer record --action ON --id 1 --metadata "zonwering-1-1"
```

Change the frequency band of the RFPlayer:

```bash
rfplayer setfreq --band H --freq 868350
rfplayer setfreq --band L --freq 433420
```

## Troubleshooting

If you need to use `sudo` to access the serial port, you can add your user to the `dialout` group:

```bash
sudo usermod -aG dialout $USER
```

Then, log out and log back in for the changes to take effect.
