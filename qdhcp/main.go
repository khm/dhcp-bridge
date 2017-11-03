package main

import (
	"encoding/json"
	"fmt"
	dhcp "github.com/krolaw/dhcp4"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// These things are necessary to construct a reply packet.
type DHCPResponse struct {
	MsgType       dhcp.MessageType `json:"msgType"`
	Server        net.IP           `json:"server"`
	ClientIP      net.IP           `json:"clientIP"`
	LeaseDuration time.Duration    `json:"leaseDuration,omitempty"`
	Options       dhcp.Options     `json:"options,omitempty"`
}

func GetIP(writer http.ResponseWriter, request *http.Request) {

	// this is where we ask a real dhcp server for a packet
	// for now we'll just default to garbage and optionally
	// provide different garbage.
	ipstring := "127.0.0.1"
	if strings.Contains(request.RequestURI, "00:00:00:00:00:01") {
		ipstring = "127.0.0.2"
	}
	ip := net.ParseIP(ipstring)

	response := DHCPResponse{
		Server:   net.ParseIP("127.0.0.1"),
		MsgType:  2, // offer
		ClientIP: ip,
	}
	// at this point "response" is populated and ready to encode

	encodedResponse, err := json.Marshal(response)
	if err != nil {
		log.Fatal(err)
	}

	stringResponse := fmt.Sprintf("%s", encodedResponse)
	fmt.Fprintf(writer, stringResponse)

}

func main() {

	http.HandleFunc("/mac/", GetIP)
	log.Fatal(http.ListenAndServe(":3001", nil))
}
