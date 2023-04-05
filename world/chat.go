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

package world

import "time"

// SetLastChatTimestamp update the lastChatTimestamp and return true if new timestamp is newer than last one.
// Otherwise, didn't update the lastChatTimestamp and return false.
func (p *Player) SetLastChatTimestamp(t time.Time) bool {
	if p.lastChatTimestamp.Before(t) {
		p.lastChatTimestamp = t
		return true
	}
	return false
}

func (p *Player) GetPrevChatSignature() []byte    { return p.lastChatSignature }
func (p *Player) SetPrevChatSignature(sig []byte) { p.lastChatSignature = sig }
