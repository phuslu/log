//go:build go1.18
// +build go1.18

package log

import (
	"net/netip"
)

// NetIPAddr adds IPv4 or IPv6 Address to the entry.
func (e *TSVEntry) NetIPAddr(ip netip.Addr) *TSVEntry {
	e.buf = ip.AppendTo(e.buf)
	e.buf = append(e.buf, e.sep)
	return e
}

// NetIPAddrPort adds IPv4 or IPv6 with Port Address to the entry.
func (e *TSVEntry) NetIPAddrPort(ipPort netip.AddrPort) *TSVEntry {
	e.buf = ipPort.AppendTo(e.buf)
	e.buf = append(e.buf, e.sep)
	return e
}

// NetIPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the entry.
func (e *TSVEntry) NetIPPrefix(pfx netip.Prefix) *TSVEntry {
	e.buf = pfx.AppendTo(e.buf)
	e.buf = append(e.buf, e.sep)
	return e
}
