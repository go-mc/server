package client

import pk "github.com/Tnze/go-mc/net/packet"

type clientInformation struct{}

func (c clientInformation) Handle(p pk.Packet, client *Client) error {
	var (
		Locale              pk.String
		ViewDistance        pk.Byte
		ChatMode            pk.VarInt
		ChatColors          pk.Boolean
		DisplayedSkinParts  pk.UnsignedByte
		MainHand            pk.VarInt
		EnableTextFiltering pk.Boolean
		AllowServerListings pk.Boolean
	)
	err := p.Scan(
		&Locale,
		&ViewDistance,
		&ChatMode,
		&ChatColors,
		&DisplayedSkinParts,
		&MainHand,
		&EnableTextFiltering,
		&AllowServerListings,
	)
	if err != nil {
		return err
	}

	//client.SetLocale(string(Locale))
	//client.SetViewDistance(int(ViewDistance))
	//client.SetChatMode(byte(ChatMode))
	//client.SetChatColors(bool(ChatColors))
	//client.SetDisplayedSkinParts(byte(DisplayedSkinParts))
	//client.SetMainHand(byte(MainHand))
	//client.SetEnableTextFiltering(bool(EnableTextFiltering))
	//client.SetAllowServerListings(bool(AllowServerListings))
	return nil
}
