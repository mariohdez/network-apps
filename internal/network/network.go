package network

type Network string

const (
	UDPNetwork Network = "udp"
	TCPNetwork Network = "tcp"
)

var (
	networkToString map[Network]string = map[Network]string{
		UDPNetwork: "udp",
		TCPNetwork: "tcp",
	}
)

func (n Network) String() string {
	return networkToString[n]
}
