package mme

import "net"

func (*Server) ListenAddrAdd(ips []net.IP) {

}

func (*Server) ListenAddrDel(ips []net.IP) {

}

// ListenAddrSet set local listen address of the MME server. When ips is empty
// slice, listen address will be cleared. That means all of local address will
// be listened.
func (*Server) ListenAddrSet(ips []net.IP) {

}
