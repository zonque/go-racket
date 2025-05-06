package rotor

import (
	"fmt"
	"net"
)

type MulticastAddress net.UDPAddr

func (m MulticastAddress) String() string {
	return fmt.Sprintf("%s:%d", m.IP, m.Port)
}
