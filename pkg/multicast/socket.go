package multicast

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/ipv4"
)

func OpenPacketConn(bindAddr net.IP, port int, ifname string) (*ipv4.PacketConn, error) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return nil, fmt.Errorf("socket syscall failed: %w", err)
	}

	if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return nil, fmt.Errorf("failed to set SO_REUSEADDR: %w", err)
	}

	// if err := syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1); err != nil {
	// 	return nil, fmt.Errorf("failed to set SO_REUSEPORT: %w", err)
	// }

	// if err := syscall.SetsockoptString(s, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, ifname); err != nil {
	// 	log.Fatal(err)
	// }

	lsa := syscall.SockaddrInet4{Port: port}
	copy(lsa.Addr[:], bindAddr.To4())
	// copy(lsa.Addr[:], []byte{0, 0, 0, 0})

	if err := syscall.Bind(s, &lsa); err != nil {
		syscall.Close(s)
		log.Fatal(err)
	}

	f := os.NewFile(uintptr(s), "")
	c, err := net.FilePacketConn(f)
	f.Close()

	if err != nil {
		log.Fatal(err)
	}

	return ipv4.NewPacketConn(c), nil
}

func OpenPacketConns(ifis []*net.Interface, port int) ([]*ipv4.PacketConn, error) {
	var pcs []*ipv4.PacketConn

	for _, ifi := range ifis {
		addrs, err := ifi.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				bindAddr := ipnet.IP.To4()

				p, err := OpenPacketConn(bindAddr, port, ifi.Name)
				if err != nil {
					return nil, err
				}

				pcs = append(pcs, p)
			}
		}
	}

	return pcs, nil
}
