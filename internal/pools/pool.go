package pools

import "net"

// AddressSpace describes a space of IP addresses and provides capability to draw
// IPs from it.
type AddressSpace interface {

	// IPForService produces an allocatable IP for service
	// The pool implementation guarantees that the IP remains valid until
	// returned to the pool
	IPForService(service string) (*net.IP, error)

	// ReturnIP returns an IP to the pool after use.
	ReturnIP(*net.IP)

	// CIDR returns the IP address range handled by this pool
	CIDR() []*net.IPNet
}
