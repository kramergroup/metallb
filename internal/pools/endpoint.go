package pools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
)

// -----------------------------------------------------------------------------
// AddressSpace implementation
// -----------------------------------------------------------------------------

// EndpointAddressSpace implements a RESTful service client for the AddressSpace
// interface.
type EndpointAddressSpace struct {
	endpoint endpoint
	Cidrs    []*net.IPNet
}

// NewEndpointAddressSpace creates a new Endpoint
func NewEndpointAddressSpace(URL, authToken string, cidrs []*net.IPNet) *EndpointAddressSpace {
	return &EndpointAddressSpace{
		endpoint: endpoint{
			URL:       URL,
			AuthToken: authToken,
		},
		Cidrs: cidrs,
	}
}

// IPForService obtains an IP from the RESTful API.
func (p *EndpointAddressSpace) IPForService(service string) (*net.IP, error) {
	ip, err := p.endpoint.ObtainIP(service)
	return &ip, err
}

// ReturnIP returns an IP to the API after use.
func (p *EndpointAddressSpace) ReturnIP(ip *net.IP) {
	p.endpoint.InvalidateIP(*ip)
}

// CIDR returns the IP address range handled by this pool
// FIXME: The implementation quietly absorbs errors from the endpoint. This is hard
// to debug. We should find a way to communicate errors
func (p *EndpointAddressSpace) CIDR() []*net.IPNet {

	return p.Cidrs
	// res, err := p.endpoint.configuration()
	// if err != nil {
	// 	return make([]*net.IPNet, 0)
	// }
	//
	// cidrs := []*net.IPNet{}
	// for _, cidr := range res.CIDRs {
	// 	if c, err := parseCIDR(cidr); err == nil {
	// 		cidrs = append(cidrs, c...)
	// 	} else {
	// 		log.Println("Error [CIDR] > %s", err.Error())
	// 	}
	// }
	// return cidrs
}

// -----------------------------------------------------------------------------
// IP service API client
// -----------------------------------------------------------------------------

// endpoint provides IP minting services and implements the newIPRequest API
type endpoint struct {
	URL       string // the url of the endpoint
	AuthToken string // authentication token send with requests (unused at the moment)
}

// newIPRequest is send to Endpoint to request minting of a new IP
type newIPRequest struct {
	Service string `json:"service"` // Name of the service the IP is intended for
}

// invalidateIPRequest is send to Endpoint to inform the service that
// this IP is no longer in use
type invalidateIPRequest struct {
	IP string `json:"ip"` // IP that is released
}

// newIPRequestResponse is send as response to newIPRequest requests
type newIPRequestResponse struct {
	IP     string `json:"ip"`
	Status string `json:"status"`
}

// configurationRequestResponse is send in response to a configuration request
// The endpoint communicates its current configuration
type configurationRequestResponse struct {
	CIDRs []string `json:"cidrs"` // Array of CIDR nets controlled by the endpoint
}

type apiEndpoint struct {
	TemplateURL string
	Method      string
}

const (
	// newIPRequestResponseStatusOK indicates a successful response to an newIPRequest
	newIPRequestResponseStatusOK = "success"

	// validateIPRequestResponseVALID indicates a valid IP
	validateIPRequestResponseVALID = "valid"
)

var (
	apiEndpointObtainIP = apiEndpoint{
		TemplateURL: "%s/v1/ip",
		Method:      "POST",
	}

	apiEndpointReturnIP = apiEndpoint{
		TemplateURL: "%s/v1/ip",
		Method:      "DELETE",
	}

	apiEndpointConfiguration = apiEndpoint{
		TemplateURL: "%s/v1/config",
		Method:      "GET",
	}
)

// ObtainIP queries the pool endpoint for a new IP to assign to services
func (e *endpoint) ObtainIP(svc string) (net.IP, error) {

	// Encode payload
	r := newIPRequest{
		Service: svc,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	// Build the request
	req, err := http.NewRequest(apiEndpointObtainIP.Method,
		fmt.Sprintf(apiEndpointObtainIP.TemplateURL, e.URL), b)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if e.AuthToken != "" {
		req.Header.Set("X-Auth-Token", e.AuthToken)
	}

	if err != nil {
		log.Print("EndpointRequest: ", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("EndpointRequest: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	var response newIPRequestResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println(err)
	}

	if response.Status != newIPRequestResponseStatusOK {
		return nil, errors.New(response.Status)
	}

	ret := net.ParseIP(response.IP)
	if ret == nil {
		return nil, fmt.Errorf("Service returned malformated IP [%s]", response.IP)
	}
	return ret, nil
}

// InvalidateIP informs the endpoint that an IP is no longer needed. Endpoints should
// take care to release resources as appropriate.
func (e *endpoint) InvalidateIP(ip net.IP) error {

	// Encode payload
	r := invalidateIPRequest{
		IP: ip.String(),
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	// Build the request
	req, err := http.NewRequest(apiEndpointReturnIP.Method,
		fmt.Sprintf(apiEndpointReturnIP.TemplateURL, e.URL), b)
	req.Header.Set("Content-Type", "application/json")
	if e.AuthToken != "" {
		req.Header.Set("X-Auth-Token", e.AuthToken)
	}

	if err != nil {
		log.Print("invalidateIPRequest: ", err)
		return err
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("invalidateIPRequest: ", err)
		return err
	}
	defer resp.Body.Close()

	// Ignore response - nothing we can do anyway if the service is not happy
	return nil
}

func (e *endpoint) configuration() (*configurationRequestResponse, error) {

	// Build the request
	req, err := http.NewRequest(apiEndpointConfiguration.Method,
		fmt.Sprintf(apiEndpointConfiguration.TemplateURL, e.URL), nil)
	req.Header.Set("Content-Type", "application/json")
	if e.AuthToken != "" {
		req.Header.Set("X-Auth-Token", e.AuthToken)
	}

	if err != nil {
		log.Print("configurationRequest: ", err)
		return nil, err
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("configurationRequest: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	//
	response := new(configurationRequestResponse)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		log.Print("validateIPRequest: ", err)
		return nil, err
	}

	return response, nil
}

/*
// ObtainLease obtains a new IP lease from a DHCP service
func ObtainLease(svc, ifname string, onBound, onExpire func(*dhclient.Lease)) {

	iface, err := net.InterfaceByName(ifname)

	if err != nil {
		fmt.Printf("unable to find interface %s: %s\n", ifname, err)
		os.Exit(1)
	}

	client := dhclient.Client{
		Iface:    iface,
		Hostname: svc,
		OnBound:  onBound,
	}

	client.Start()
}
*/
