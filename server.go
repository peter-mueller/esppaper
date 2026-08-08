package main

import (
	"net/netip"

	"github.com/soypat/lneto/http/httphi"
	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/netlink"
	esplink "tinygo.org/x/espradio/netlink"
)

type NetworkLink struct {
	SSID       string
	Passphrase string

	l *esplink.Esplink
}

func (nl *NetworkLink) Address() (netip.Addr, error) {
	return nl.l.Addr()
}

func (nl *NetworkLink) ListenAndServe(r *httphi.Router, port string) error {
	return nl.ListenAndServe(r, port)
}

func (nl *NetworkLink) Connect() error {
	l := &esplink.Esplink{}
	netdev.UseNetdev(l)

	err := l.NetConnect(&netlink.ConnectParams{
		Ssid:       nl.SSID,
		Passphrase: nl.Passphrase,
	})
	if err != nil {
		return err
	}
	nl.l = l
	return nil
}
