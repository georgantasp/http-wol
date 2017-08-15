package main

import (
  "bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
  "net/http"
  "os"
	"regexp"
)

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", sendWol)
  http.ListenAndServe(":" + os.Args[1], mux)
}

func sendWol(w http.ResponseWriter, r *http.Request) {
  macAddr := r.URL.Query().Get("mac")
  
  err := wol.SendMagicPacket(macAddr, "255.255.255.255:9", "")
  
  if err != nil {
    fmt.Fprintf(w, "ERROR: %s\n", err.Error())
  } else {
    fmt.Fprintf(w, "Magic packet sent successfully to %s\n", macAddr)
  }
}

// from https://github.com/sabhiram/go-wol

// Define globals for the MacAddress parsing
var (
	delims = ":-"
	re_MAC = regexp.MustCompile(`^([0-9a-fA-F]{2}[` + delims + `]){5}([0-9a-fA-F]{2})$`)
)

type MACAddress [6]byte

// A MagicPacket is constituted of 6 bytes of 0xFF followed by
// 16 groups of the destination MAC address.
type MagicPacket struct {
	header  [6]byte
	payload [16]MACAddress
}

// This function accepts a MAC Address string, and returns a pointer to
// a MagicPacket object. A Magic Packet is a broadcast frame which
// contains 6 bytes of 0xFF followed by 16 repetitions of a given mac address.
func NewMagicPacket(mac string) (*MagicPacket, error) {
	var packet MagicPacket
	var macAddr MACAddress

	// We only support 6 byte MAC addresses since it is much harder to use
	// the binary.Write(...) interface when the size of the MagicPacket is
	// dynamic.
	if !re_MAC.MatchString(mac) {
		return nil, errors.New("MAC address " + mac + " is not valid.")
	}

	hwAddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}

	// Copy bytes from the returned HardwareAddr -> a fixed size
	// MACAddress
	for idx := range macAddr {
		macAddr[idx] = hwAddr[idx]
	}

	// Setup the header which is 6 repetitions of 0xFF
	for idx := range packet.header {
		packet.header[idx] = 0xFF
	}

	// Setup the payload which is 16 repetitions of the MAC addr
	for idx := range packet.payload {
		packet.payload[idx] = macAddr
	}

	return &packet, nil
}

// Function to send a magic packet to a given mac address, and optionally
// receives an iface to broadcast on. An iface of "" implies a nil net.UDPAddr
func SendMagicPacket(macAddr, bcastAddr, iface string) error {
	// Construct a MagicPacket for the given MAC Address
	magicPacket, err := NewMagicPacket(macAddr)
	if err != nil {
		return err
	}

	// Fill our byte buffer with the bytes in our MagicPacket
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, magicPacket)

	// Get a UDPAddr to send the broadcast to
	udpAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		fmt.Printf("Unable to get a UDP address for %s\n", bcastAddr)
		return err
	}

	// Open a UDP connection, and defer it's cleanup
	connection, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return errors.New("Unable to dial UDP address.")
	}
	defer connection.Close()

	// Write the bytes of the MagicPacket to the connection
	bytesWritten, err := connection.Write(buf.Bytes())
	if err != nil {
		fmt.Printf("Unable to write packet to connection\n")
		return err
	} else if bytesWritten != 102 {
		fmt.Printf("Warning: %d bytes written, %d expected!\n", bytesWritten, 102)
	}

	return nil
}
