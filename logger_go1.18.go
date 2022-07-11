//go:build go1.18
// +build go1.18

package log

import (
	"net/netip"
)

// NetIPAddr adds IPv4 or IPv6 Address to the entry.
func (e *Entry) NetIPAddr(key string, ip netip.Addr) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = ip.AppendTo(e.buf)
	e.buf = append(e.buf, '"')
	return e
}

// NetIPAddrPort adds IPv4 or IPv6 with Port Address to the entry.
func (e *Entry) NetIPAddrPort(key string, ipPort netip.AddrPort) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = ipPort.AppendTo(e.buf)
	e.buf = append(e.buf, '"')
	return e
}

// NetIPPrefix adds IPv4 or IPv6 Prefix (address and mask) to the entry.
func (e *Entry) NetIPPrefix(key string, pfx netip.Prefix) *Entry {
	if e == nil {
		return nil
	}
	e.buf = append(e.buf, ',', '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':', '"')
	e.buf = pfx.AppendTo(e.buf)
	e.buf = append(e.buf, '"')
	return e
}
