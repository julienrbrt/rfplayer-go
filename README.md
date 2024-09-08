# RFPlayer-Go

CLI tool and API for [RFPlayer](https://www.gce-electronics.com/fr/produits-radio/1777-rf-player-3770008041004.html), in Go.
This CLI tool allows to configure the RFPlayer and to send commands to it via the CLI or via an API.

## Goals

The idea is to configure to use the tool locally via its CLI, or on [GoKrazy](https://gokrazy.dev/), to send commands to the RFPlayer to control or record RF devices. The goal is to support HomeKit via [Hap](https://github.com/brutella/hap) and be able to control RF devices via HomeKit. This project is still a **work in progress**.

## Installation

```bash
go install github.com/julienrbrt/rfplayer-go/cmd/rfplayer@latest
```

## How to use

```bash
rfplayer --help
```
