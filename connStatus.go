package main

import (
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
	dlySec        int
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

// TODO: init from command line parsing
func initParameters() (parameters connectionParameters) {
	parameters = connectionParameters{
		listenNetwork: "udp4",
		listenAddress: "0.0.0.0",
		mtu:           1500,
		peer:          "75.75.75.75",
		iface:         "en0",
		count:         2,
		dlySec:        1,
	}
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
		case ipv4.ICMPTypeEchoReply:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeDestinationUnreachable:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeRedirect:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeEcho:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeRouterAdvertisement:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeRouterSolicitation:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeTimeExceeded:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeParameterProblem:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeTimestamp:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeTimestampReply:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypePhoturis:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeExtendedEchoRequest:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		case ipv4.ICMPTypeExtendedEchoReply:
			fmt.Printf("got %v from %v\n", rm.Type, peer)
		default:
			fmt.Printf("got %+v; unknown\n", rm.Type)
		}
	}
}
