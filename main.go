package main

import (
	"log/slog"
	"machine"
	"os"
	"time"

	"github.com/peter-mueller/esppaper/crowpanel579"
	"github.com/peter-mueller/esppaper/networklink"
	"github.com/soypat/lneto/http/httphi"
)

var (
	wlanSSID       string
	wlanPassphrase string
	hostname       string
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
		machine.GPIO7,  // Power
		machine.GPIO45, // Chip Select 1
		machine.GPIO46, // Data/Command Control
		machine.GPIO47, // Hardware Reset
		machine.GPIO48, // Busy
	)
}

func main() {
	slog.Info("Starting Device")

	device := Device{
		Epaper: NewEpaper(),
	}
	device.Epaper.Power(true)
	device.Epaper.Configure()
	device.Epaper.Power(false)

	networkLink := networklink.NetworkLink{
		SSID:       wlanSSID,
		Passphrase: wlanPassphrase,
		Hostname:   hostname,
	}

	err := networkLink.Connect()
	mustNil("Failed to connect NetworkLink", err)

	addr, err := networkLink.Address()
	mustNil("No NetworkLink Address", err)
	slog.Info("Connected to Wifi", "SSID", networkLink.SSID, "addr", addr.String())

	var http httphi.MuxSlice
	http.Handle("POST /epaper", device.postEpaper)

	var router httphi.Router
	cfg := httphi.DefaultRouterConfig(1, 2048, http.MaxPathValues())
	mustNil("configuring Router", router.Configure(&http, cfg))
	mustNil("listen and serve", networkLink.ListenAndServe(&router, 80))
}

func (device *Device) postEpaper(exch *httphi.Exchange) {
	size, present, err := exch.RequestContentLength()
	if err != nil {
		slog.Error("failed to read Content-Legth", "err", err)
		exch.WriteHeader(httphi.StatusInternalServerError)
		return
	}
	if !present {
		exch.WriteBodyString("Content-Length must be present")
		exch.WriteHeader(httphi.StatusBadGateway)
		return
	}
	body := make([]byte, size)
	exch.ReadBody(body)
	slog.Info("Displaying Image with Text", "size", size)

	img := DrawText(string(body), 12)
	exch.WriteHeader(httphi.StatusOK)

	device.displayImage(img)
}

func (d *Device) displayImage(img *Image) {
	d.Epaper.Power(true)
	defer d.Epaper.Power(false)

	time.Sleep(10 * time.Millisecond)

	bounds := img.Bounds()
	for x := 0; x < crowpanel579.Width && x < bounds.Max.X; x++ {
		for y := 0; y < crowpanel579.Height && y < bounds.Max.Y; y++ {
			d.Epaper.SetPixel(int16(x), int16(y), img.GrayAt(x, y))
		}
	}
	d.Epaper.Display()
	time.Sleep(100 * time.Millisecond)

}

func mustNil(msg string, v error) {
	if v != nil {
		slog.Error(msg, "err", v.Error())
		os.Exit(1)
	}
}
