package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"runtime"
)

type connectionParameters struct {
	listenNetwork string
	listenAddress string
	mtu           int
	peer          string
	iface         string
	count         int
	delaySec      int
}

func main() {

	if unprivilegedICMP() {
		parameters := initParameters()
		ConnStatus(&parameters)
	}
}

// Check if this OS supports unprivileged ICMP.
func unprivilegedICMP() (retcode bool) {
	retcode = true
	switch runtime.GOOS {
	case "darwin": // yes
		fmt.Println("unprivileged ICMP enabled")
	case "linux": // maybe?
		fmt.Println("you may need to adjust the net.ipv4.ping_group_range kernel state")
	default: // nope
		fmt.Println("not supported on", runtime.GOOS)
		retcode = false
	}
	return
}

func initParameters() (parameters connectionParameters) {
	parameters = connectionParameters{
		listenNetwork: "udp4",
		listenAddress: "0.0.0.0",
		mtu:           1500,
		peer:          "",
		iface:         "en0",
		count:         0,
		delaySec:      0,
	}
	flag.StringVar(&parameters.peer, "peer", "75.75.75.75", "ping target ipv4 address")
	flag.IntVar(&parameters.count, "count", 1, "ping repeat count")
	flag.IntVar(&parameters.delaySec, "delaySec", 1, "delay in seconds between pings")
	flag.Parse()
	return
}

func ConnStatus(parameters *connectionParameters) {

	connection, err := icmp.ListenPacket(parameters.listenNetwork, parameters.listenAddress)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = connection.Close()
		if err != nil {
			fmt.Printf("%v\n", err)
		}
	}()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("hello-sailor"),
		},
	}
	msgTx, err := msg.Marshal(nil)
	if err != nil {
		panic(err)
	}

	msgRx := make([]byte, parameters.mtu)
	udpAddr := net.UDPAddr{IP: net.ParseIP(parameters.peer), Zone: parameters.iface}

	for i := 0; i < parameters.count; i++ {

		if _, err := connection.WriteTo(msgTx, &udpAddr); err != nil {
			fmt.Println(err)
		}

		n, peer, err := connection.ReadFrom(msgRx)
		if err != nil {
			fmt.Println(err)
		}
		rm, err := icmp.ParseMessage(1, msgRx[:n])
		if err != nil {
			fmt.Println(err)
		}

		switch rm.Type {
		case
			ipv4.ICMPTypeEchoReply,
			ipv4.ICMPTypeDestinationUnreachable,
			ipv4.ICMPTypeRedirect,
			ipv4.ICMPTypeEcho,
			ipv4.ICMPTypeRouterAdvertisement,
			ipv4.ICMPTypeRouterSolicitation,
			ipv4.ICMPTypeTimeExceeded,
			ipv4.ICMPTypeParameterProblem,
			ipv4.ICMPTypeTimestamp,
			ipv4.ICMPTypeTimestampReply,
			ipv4.ICMPTypePhoturis,
			ipv4.ICMPTypeExtendedEchoRequest,
			ipv4.ICMPTypeExtendedEchoReply:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		default:
			fmt.Printf("got %v; unknown\n", rm.Type)
		}
	}
}
