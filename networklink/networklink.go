package networklink

import (
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/soypat/lneto/http/httphi"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	esplink "tinygo.org/x/espradio/netlink"
)

type NetworkLink struct {
	SSID       string
	Passphrase string
	Hostname   string

	l *esplink.Esplink
}

func (nl *NetworkLink) Address() (netip.Addr, error) {
	return nl.l.Addr()
}

func (nl *NetworkLink) ListenAndServe(r *httphi.Router, port uint16) error {
	return nl.l.ListenAndServe(r, port)
}

func (nl *NetworkLink) Connect() error {
	l := &esplink.Esplink{}
	netdev.UseNetdev(l)

	const attempts = 3
	for range attempts {
		err := l.NetConnect(&netlink.ConnectParams{
			Ssid:       nl.SSID,
			Passphrase: nl.Passphrase,
			Hostname:   nl.Hostname,
		})
		if err == nil {
			nl.l = l
			return nil
		}
		if err.Error() == "espradio: radio already enabled" {
			nl.l = l
			return nil
		}
		slog.Warn("failed to connect to WiFi", "err", err)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("failed to connect to WiFi after %d attemts", attempts)
}
