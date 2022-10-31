package network

import (
	"fmt"
	"net"

	"github.com/prometheus/procfs"
)

// Ref: https://unix.stackexchange.com/a/470527
type TCPConnStatus uint64

const (
	TCPConnStatusEstablished TCPConnStatus = 1
	TCPConnStatusListening   TCPConnStatus = 10
)

func GetOpenedTCPConns() (procfs.NetTCP, error) {
	proc, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("could not read /proc: %s", err)
	}

	tcpIPv4, err := proc.NetTCP()
	if err != nil {
		return nil, fmt.Errorf("could not read /proc/net/tcp: %s", err)
	}

	tcpIPv6, err := proc.NetTCP6()
	if err != nil {
		return nil, fmt.Errorf("could not read /proc/net/tcp6: %s", err)
	}

	return append(tcpIPv4, tcpIPv6...), nil
}

func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
