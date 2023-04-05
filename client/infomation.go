// This file is part of go-mc/server project.
// Copyright (C) 2023.  Tnze
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
