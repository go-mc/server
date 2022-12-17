package client

import (
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/world"
)

type clientInformation struct{}

func (c clientInformation) Handle(p pk.Packet, client *Client) error {
	var info world.ClientInfo
	if err := p.Scan(&info); err != nil {
		return err
	}
	client.Inputs.Lock()
	client.Inputs.ClientInfo = info
	client.Inputs.Unlock()
	return nil
}
