package addressmonitor

import (
	"context"
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

type AddressMonitor struct {
	Ch chan []net.IPNet
}

func New(ctx context.Context, deviceName string) (*AddressMonitor, error) {
	m := new(AddressMonitor)
	m.Ch = make(chan []net.IPNet, 1)
	updateCh := make(chan netlink.AddrUpdate)
	doneCh := make(chan struct{})
	addresses := []net.IPNet{}

	err := netlink.AddrSubscribe(updateCh, doneCh)
	if err != nil {
		return nil, fmt.Errorf("can not subscribe to address changes: %w", err)
	}

	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		return nil, fmt.Errorf("can not get link: %w", err)
	}

	list, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("can not get address list: %w", err)
	}

	for _, addr := range list {
		addresses = append(addresses, *addr.IPNet)
	}

	m.Ch <- addresses

	go func() {
		for {
			select {
			case u := <-updateCh:
				link, err := netlink.LinkByIndex(u.LinkIndex)
				if err != nil {
					continue
				}

				if link.Attrs().Name != deviceName {
					continue
				}

				if u.NewAddr {
					exists := false

					for _, a := range addresses {
						if a.IP.Equal(u.LinkAddress.IP) {
							exists = true
						}
					}

					if !exists {
						addresses = append(addresses, u.LinkAddress)
					}
				} else {
					for i, a := range addresses {
						if a.IP.Equal(u.LinkAddress.IP) {
							addresses = append(addresses[:i], addresses[i+1:]...)
						}
					}
				}

				m.Ch <- addresses

			case <-ctx.Done():
				close(doneCh)

				return
			}
		}
	}()

	return m, nil
}
