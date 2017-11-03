package main

import (
	dhcp "github.com/krolaw/dhcp4"

	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// These things are necessary to construct a reply packet.
type DHCPResponse struct {
	Packet        dhcp.Packet      `json:"packet"`
	MsgType       dhcp.MessageType `json:"msgType"`
	Server        net.IP           `json:"server"`
	ClientIP      net.IP           `json:"clientIP"`
	LeaseDuration time.Duration    `json:"leaseDuration"`
	Options       dhcp.Options     `json:"options"`
}

func getDHCPResponse(api url.URL, key string) (response DHCPResponse) {
	var info DHCPResponse

	apistring := fmt.Sprintf("%s", api)

	query := apistring + key
	log.Println("Requesting", query)

	httpResponse, err := http.Get(query)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		log.Fatal(err)
	}

	return info
}

type DHCPHandler struct {
	srv net.IP
	api url.URL
}

func (han *DHCPHandler) ServeDHCP(req dhcp.Packet, msg dhcp.MessageType, reqopts dhcp.Options) (d dhcp.Packet) {
	reqnic := req.CHAddr().String()
	log.Println("got", msg, "from", reqnic)

	// initialize a safe and useless reply
	var reply DHCPResponse

	reply.Packet = req
	reply.MsgType = dhcp.NAK
	reply.Server = han.srv
	reply.ClientIP = nil
	reply.LeaseDuration = 0
	reply.Options = nil

	switch msg {

	case dhcp.Discover:
		reply = getDHCPResponse(han.api, reqnic)
		reply.MsgType = dhcp.Offer

	case dhcp.Request:
		if !net.IP(reqopts[dhcp.OptionServerIdentifier]).Equal(han.srv) {
			return nil // wasn't asking us
		}
		reply = getDHCPResponse(han.api, reqnic)
		reply.MsgType = dhcp.ACK

	default:
		// includes Release, Decline, Inform msg types
		// no response packet needed
		// TODO:
		// send reqnic upstream
		// for Release and Decline also signify lease term
		return nil
	}

	return dhcp.ReplyPacket(reply.Packet,
		reply.MsgType,
		reply.Server,
		reply.ClientIP,
		reply.LeaseDuration,
		reply.Options.SelectOrderOrAll(nil))

}

func main() {
	ipFlag := flag.String("ip", "0.0.0.0", "listen address")
	apiFlag := flag.String("api", "http://localhost/mac/", "api endpoint")

	flag.Parse()

	ip := net.ParseIP(*ipFlag)
	if ip == nil {
		log.Fatal("invalid listen IP address specified")
	}

	apiURL, err := url.Parse(*apiFlag)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Starting up on", ip, "serving requests from", apiURL)

	handler := &DHCPHandler{
		srv: ip,
		api: *apiURL,
	}

	log.Fatal(dhcp.ListenAndServe(handler))

}
