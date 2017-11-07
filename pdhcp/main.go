package main

import (
	dhcp "github.com/krolaw/dhcp4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	statsRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dhcp_requests",
		Help: "The number of requests processed.",
	});
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

func getDHCPResponse(api url.URL, key string) (response *DHCPResponse) {
	var info *DHCPResponse

	apistring := api.String()

	query := apistring + key
	log.Println("Requesting", query)

	httpResponse, err := http.Get(query)
	if err != nil {
		log.Print(err)
		return nil
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		log.Print(err)
		return nil
	}

	err = json.Unmarshal(body, info)
	if err != nil {
		log.Print(err)
		return nil
	}

	return info
}

type DHCPHandler struct {
	srv net.IP
	api url.URL
}

func (han *DHCPHandler) ServeDHCP(req dhcp.Packet, msg dhcp.MessageType, reqopts dhcp.Options) (d dhcp.Packet) {
	statsRequests.Inc()
	reqnic := req.CHAddr().String()
	log.Println("got", msg, "from", reqnic)

	// initialize a safe and useless reply
	var reply *DHCPResponse

	switch msg {

	case dhcp.Discover:
		reply = getDHCPResponse(han.api, reqnic)

	case dhcp.Request:
		if !net.IP(reqopts[dhcp.OptionServerIdentifier]).Equal(han.srv) {
			return nil // wasn't asking us
		}
		reply = getDHCPResponse(han.api, reqnic)

	default:
		// includes Release, Decline, Inform msg types
		// no response packet needed
		// TODO:
		// send reqnic upstream
		// for Release and Decline also signify lease term
		return nil
	}

	if reply == nil {
		return nil
	}

	return dhcp.ReplyPacket(reply.Packet,
		reply.MsgType,
		reply.Server,
		reply.ClientIP,
		reply.LeaseDuration,
		reply.Options.SelectOrderOrAll(nil))

}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(statsRequests)
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

	// Expose the registered metrics via HTTP.
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("0.0.0.0:9000", nil))
	}()

	handler := &DHCPHandler{
		srv: ip,
		api: *apiURL,
	}

	log.Fatal(dhcp.ListenAndServe(handler))

}
