package ip

type Ip interface {
	IsIpv4() bool
}
type ipRange interface {
	pickup() []*Ip
}
