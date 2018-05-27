package pools

import (
	"errors"
	"log"
	"net"

	"github.com/mikioh/ipaddr"
)

// CIDRAddressSpace implements a simple AddressSpace based on provided
// address ranges.
type CIDRAddressSpace struct {
	deployed *set
	cidrs    []*net.IPNet
}

// NewFixedCIDRAddressSpace creates a new CIDRAddressSpace from the given
// array of CIDRs
func NewFixedCIDRAddressSpace(cidrs []string) *CIDRAddressSpace {

	pool := &CIDRAddressSpace{
		cidrs:    make([]*net.IPNet, 0),
		deployed: newSet(),
	}
	for _, cidr := range cidrs {
		if p, err := parseCIDR(cidr); err == nil {
			pool.cidrs = append(pool.cidrs, p...)
		} else {
			log.Printf("Error parsing CIDR: %s", err.Error())
		}
	}
	return pool
}

// -----------------------------------------------------------------------------

// IPForService returns a new IP for service svc
func (p *CIDRAddressSpace) IPForService(svc string) (*net.IP, error) {

	for _, cidr := range p.CIDR() {
		c := ipaddr.NewCursor([]ipaddr.Prefix{*ipaddr.NewPrefix(cidr)})
		for pos := c.First(); pos != nil; pos = c.Next() {
			ip := pos.IP
			// Somewhat inefficiently brute-force by checking all deployed IPs
			if !p.deployed.Contains(&ip) {
				p.deployed.Add(&ip)
				return &ip, nil
			}
		}
	}

	return nil, errors.New("No available IPs")
}

// ReturnIP returns an IP to the AddressSpace after use
func (p *CIDRAddressSpace) ReturnIP(ip *net.IP) {
	p.deployed.Remove(ip)
}

// CIDR returns an array of IP networks covered by the AddressSpace
func (p *CIDRAddressSpace) CIDR() []*net.IPNet {
	return p.cidrs
}
