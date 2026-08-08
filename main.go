package main

import (
	"image/color"
	"log/slog"
	"machine"
	"os"
	"time"

	"github.com/peter-mueller/esppaper/crowpanel579"
)

var (
	wlanSSID       string
	wlanPassphrase string
	serverPort     string = "80"
)

type Device struct {
	Epaper crowpanel579.CrowPanel579
}

func NewEpaper() crowpanel579.CrowPanel579 {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 2 * 1000 * 1000,
		SCK:       machine.GPIO12,
		SDO:       machine.GPIO11,
	})

	// 1. Configure the shared SPI bus
	return crowpanel579.NewCrowPanel579(
		machine.SPI0,
		machine.GPIO45, // Chip Select 1
		machine.GPIO46, // Data/Command Control
		machine.GPIO47, // Hardware Reset
		machine.GPIO48, // Busy
	)
}

func main() {
	slog.Info("Starting Device")

	pwrPin := machine.GPIO7
	pwrPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pwrPin.High()
	time.Sleep(100 * time.Millisecond)

	device := Device{
		Epaper: NewEpaper(),
	}

	device.Epaper.Configure()
	for x := int16(0); x < crowpanel579.Width; x++ {
		for y := int16(0); y < crowpanel579.Height; y++ {
			c := color.Gray{}
			if y%8 < 4 {
				c = color.Gray{0xFF}
			}
			device.Epaper.SetPixel(x, y, c)
		}
	}
	device.Epaper.Display()

	time.Sleep(5 * time.Second)

	/**
		networkLink := NetworkLink{
			SSID:       wlanSSID,
			Passphrase: wlanPassphrase,
		}

		err := networkLink.Connect()
		mustNil("Connect NetworkLink", err)
		slog.Info("Connected to Wifi", "SSID", networkLink.SSID)

		var mux httphi.MuxSlice
		mux.Handle("/", func(exch *httphi.Exchange) {
			exch.RespondString(200, "application/json", `{"message":"hello"}`)
		})
		mux.Handle("POST /epaper", device.PostEpaper)
		var router httphi.Router
		cfg := httphi.DefaultRouterConfig(4, 2048, 8)
		err = router.Configure(&mux, cfg)
		mustNil("Configure Router", err)
		defer router.Shutdown()

		addr, err := networkLink.Address()
		mustNil("NetworkLink Address", err)

		slog.Info("Starting Webserver", "addr", addr.String(), "port", serverPort)
		err = networkLink.ListenAndServe(&router, serverPort)
		mustNil("NetworkLink ListenAndServe", err)
	**/
}

/*
func (d *Device) PostEpaper(ex *httphi.Exchange) {
	slog.Info("POST /epaper")
	var bin []byte
	ex.ReadBody(bin)
}
*/

func mustNil(msg string, v error) {
	if v != nil {
		slog.Error(msg, "err", v.Error())
		os.Exit(1)
	}
}
