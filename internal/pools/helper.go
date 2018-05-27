package pools

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/mikioh/ipaddr"
)

// -----------------------------------------------------------------------------
// Implementation of a set
// -----------------------------------------------------------------------------
type set struct {
	lock sync.Mutex
	s    map[string]*net.IP
}

func newSet() *set {
	return &set{sync.Mutex{}, make(map[string]*net.IP, 0)}
}

func (s *set) Add(i *net.IP) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, found := s.s[i.String()]
	s.s[i.String()] = i
	return !found //False if it existed already
}

func (s *set) Contains(i *net.IP) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	k := i.String()
	_, found := s.s[k]
	return found //False if it existed already
}

func (s *set) Remove(i *net.IP) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	k := i.String()
	_, found := s.s[k]
	delete(s.s, k)
	return found
}

// -----------------------------------------------------------------------------
// CIDR supporting functions
// -----------------------------------------------------------------------------
func parseCIDR(cidr string) ([]*net.IPNet, error) {
	var ret []*net.IPNet

	if !strings.Contains(cidr, "-") {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q", cidr)
		}
		ret = []*net.IPNet{n}
	} else {

		fs := strings.SplitN(cidr, "-", 2)
		if len(fs) != 2 {
			return nil, fmt.Errorf("invalid IP range %q", cidr)
		}
		start := net.ParseIP(fs[0])
		if start == nil {
			return nil, fmt.Errorf("invalid IP range %q: invalid start IP %q", cidr, fs[0])
		}
		end := net.ParseIP(fs[1])
		if end == nil {
			return nil, fmt.Errorf("invalid IP range %q: invalid end IP %q", cidr, fs[1])
		}

		for _, pfx := range ipaddr.Summarize(start, end) {
			n := &net.IPNet{
				IP:   pfx.IP,
				Mask: pfx.Mask,
			}
			ret = append(ret, n)
		}
	}

	return ret, nil
}
