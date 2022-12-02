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
