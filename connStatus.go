package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"runtime"
	"time"
)

type connectionParameters struct {
	listenNetwork string
	listenAddress string
	mtu           int
	peer          string
	iface         string
	count         int
	delaySec      int
	durationSec   time.Duration
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

// Initialize parameters with command line args.
func initParameters() (parameters connectionParameters) {
	parameters = connectionParameters{
		listenNetwork: "udp4",    // protocol
		listenAddress: "0.0.0.0", // ICMP listener bound address
		mtu:           1500,      // default
		peer:          "",        // ping target ipv4 dotted decimal address
		iface:         "en0",     // ping-er interface name
		count:         0,         // how many times to ping peer
		delaySec:      0,         // with this delay (seconds) between them
	}
	flag.StringVar(&parameters.peer, "peer", "75.75.75.75", "ping target ipv4 address")
	flag.IntVar(&parameters.count, "count", 1, "ping repeat count")
	flag.IntVar(&parameters.delaySec, "delaySec", 1, "delay in seconds between pings")
	// TODO: error checking on cl args
	flag.Parse()
	parameters.durationSec = time.Duration(parameters.delaySec) * time.Second
	return
}

// Create outgoing ping message bytes.
func createTxMsg(sequenceNumber int) []byte {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  sequenceNumber,
			Data: []byte("hello-sailor"),
		},
	}
	msgTx, err := msg.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return msgTx
}

// print timestamped connection error
func printErr(src string, err error) {
	now := time.Now().Format(time.RFC3339)
	fmt.Printf("%v connection.%s(): %v\n", now, src, err)
}

func ConnStatus(parameters *connectionParameters) {

	// create ICMP listener
	connection, err := icmp.ListenPacket(parameters.listenNetwork, parameters.listenAddress)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = connection.Close()
		if err != nil {
			printErr("Close", err)
		}
	}()

	msgRx := make([]byte, parameters.mtu)                                            // receiver data buffer
	udpAddr := net.UDPAddr{IP: net.ParseIP(parameters.peer), Zone: parameters.iface} // ping target address ino

	//
	for i := 0; i < parameters.count; i++ {

		// delay between pings
		if i > 0 {
			time.Sleep(parameters.durationSec)
		}

		// create and output next sequenced ICMP echo message to transmit
		msgTx := createTxMsg(i)
		if _, err := connection.WriteTo(msgTx, &udpAddr); err != nil {
			printErr("WriteTo", err)
			continue
		}

		// set read operation timeout
		if err = connection.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			printErr("SetReadDeadline", err)
			continue
		}

		// receive ping peer response if no timeout
		n, peer, err := connection.ReadFrom(msgRx)
		if err != nil {
			printErr("ReadFrom", err)
			continue

		} else { // got response, parse the bytes

			rm, err := icmp.ParseMessage(1, msgRx[:n])
			if err != nil {
				printErr("ParseMessage", err)
				continue
			}

			// got valid response, interpret the received message type field
			switch rm.Type {
			case ipv4.ICMPTypeEchoReply:
				body, ok := rm.Body.(*icmp.Echo)
				if ok {
					now := time.Now().Format(time.RFC3339)
					fmt.Printf("%v received %v %v from %v\n", now, rm.Type, body.Seq, peer)
				} else {
					fmt.Printf("Unable to cast interface icmp.MessageBody to struct icmp.Body\n")
				}
			case
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
				fmt.Printf("received unexpected response %v from %v\n", rm.Type, peer)
			default:
				fmt.Printf("received unknown response %v\n", rm.Type)
			}
		}
	}
}
