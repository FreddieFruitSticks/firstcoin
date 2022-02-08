package peer

//get all Peers when coming online
// start with peer 8080

type Peers struct {
	Hostnames map[string]string
	ThisHost  string
}

func NewPeers() *Peers {
	hostnames := make(map[string]string)
	// hostnames[seedHost] = seedHost
	return &Peers{
		Hostnames: hostnames,
	}
}

func (p *Peers) AddHostname(hostname string) {
	p.Hostnames[hostname] = hostname
}

func (p *Peers) RemoveHostname(hostname string) {
	delete(p.Hostnames, hostname)
}
