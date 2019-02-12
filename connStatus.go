package main

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"runtime"
)

func main() {
	ConnStatus()
}

func ConnStatus() {

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("unprivileged ICMP enabled")
	case "linux":
		fmt.Println("you may need to adjust the net.ipv4.ping_group_range kernel state")
	default:
		fmt.Println("not supported on", runtime.GOOS)
		return
	}

	c, err := icmp.ListenPacket("udp4", "0.0.0.0")
	if err != nil {
		fmt.Println(err)
	}
	defer c.Close()

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}
	wb, err := wm.Marshal(nil)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := c.WriteTo(wb, &net.UDPAddr{IP: net.ParseIP("75.75.75.75"), Zone: "en0"}); err != nil {
		fmt.Println(err)
	}

	rb := make([]byte, 1500)
	n, peer, err := c.ReadFrom(rb)
	if err != nil {
		fmt.Println(err)
	}
	rm, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		fmt.Println(err)
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeDestinationUnreachable:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeRedirect:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeEcho:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeRouterAdvertisement:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeRouterSolicitation:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeTimeExceeded:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeParameterProblem:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeTimestamp:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeTimestampReply:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypePhoturis:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeExtendedEchoRequest:
		fmt.Printf("got %v from %v", rm.Type, peer)
	case ipv4.ICMPTypeExtendedEchoReply:
		fmt.Printf("got %v from %v", rm.Type, peer)
	default:
		fmt.Printf("got %+v; unknown", rm.Type)
	}
}
