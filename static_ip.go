package checkly

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
)

// StaticIP is one of the public IP addresses or ranges that Checkly runs
// checks from, as returned by GetStaticIPs. Region is a location code such as
// "us-east-1". Address is an IPv6 range, or an IPv4 address as a /32 prefix.
// See https://www.checklyhq.com/docs/monitoring/allowlisting/#ip-range-allowlisting.
type StaticIP struct {
	Region  string
	Address netip.Prefix
}

func (c *client) GetStaticIPs(
	ctx context.Context,
) ([]StaticIP, error) {
	var IPs []StaticIP

	// getting IPv6 first
	var datav6 map[string][]string
	err := c.callPublicJSON(ctx, http.MethodGet, "/v1/static-ipv6s-by-region", &datav6, http.StatusOK)
	if err != nil {
		return nil, err
	}

	for region, ips := range datav6 {
		for _, ip := range ips {
			addr, err := netip.ParsePrefix(ip)
			if err != nil {
				return nil, fmt.Errorf("could not parse CIDR from %s: %w", ip, err)
			}

			IPs = append(IPs, StaticIP{Region: region, Address: addr})
		}
	}

	// and then IPv4
	var datav4 map[string][]string
	err = c.callPublicJSON(ctx, http.MethodGet, "/v1/static-ips-by-region", &datav4, http.StatusOK)
	if err != nil {
		return nil, err
	}

	for region, ips := range datav4 {
		for _, ip := range ips {
			addr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, fmt.Errorf("could not parse IP address from %s: %w", ip, err)
			}
			IPs = append(IPs, StaticIP{Region: region, Address: netip.PrefixFrom(addr, 32)})
		}
	}

	return IPs, nil
}
