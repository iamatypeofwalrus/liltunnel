package liltunnel

// Tunneler abstracts the different protocols such as TCP and HTTP tunnels
type Tunneler interface {
	Tunnel() error
}
