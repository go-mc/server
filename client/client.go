package client

import (
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/server"
	"github.com/go-mc/server/player"
	"github.com/go-mc/server/world"
	"go.uber.org/zap"
)

type Client struct {
	log      *zap.Logger
	conn     *net.Conn
	player   *player.Player
	queue    *server.PacketQueue
	handlers []packetHandler
}

type packetHandler interface {
	Handle(p pk.Packet, c *Client) error
}

func New(log *zap.Logger, conn *net.Conn, player *player.Player) *Client {
	return &Client{
		log:      log,
		conn:     conn,
		player:   player,
		queue:    server.NewPacketQueue(),
		handlers: defaultHandlers,
	}
}

func (c *Client) Spawn(w *world.World) error {
	err := c.SendLogin(w, c.player)
	w.AddPlayer(c, c.player)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Start() {
	stopped := make(chan struct{}, 2)
	done := func() {
		stopped <- struct{}{}
	}
	// Exit when any error is thrown
	go c.startSend(done)
	go c.startReceive(done)
	<-stopped
}

func (c *Client) startSend(done func()) {
	defer done()
	for {
		p, ok := c.queue.Pull()
		if !ok {
			return
		}
		err := c.conn.WritePacket(p)
		if err != nil {
			c.log.Debug("Send packet fail", zap.Error(err))
			return
		}
	}
}

func (c *Client) startReceive(done func()) {
	defer done()
	var packet pk.Packet
	for {
		err := c.conn.ReadPacket(&packet)
		if err != nil {
			c.log.Debug("Receive packet fail", zap.Error(err))
			return
		}
		if packet.ID < 0 || packet.ID >= int32(len(c.handlers)) {
			c.log.Debug("Invalid packet id", zap.Int32("id", packet.ID), zap.Int("len", len(packet.Data)))
			return
		}
		if handler := c.handlers[packet.ID]; handler != nil {
			err = handler.Handle(packet, c)
			if err != nil {
				c.log.Error("Handle packet error", zap.Int32("id", packet.ID), zap.Error(err))
				return
			}
		}
	}
}

func (c *Client) AddHandler(id int32, handler packetHandler) {
	c.handlers[id] = handler
}

var defaultHandlers = []packetHandler{
	packetid.ServerboundAcceptTeleportation:      nil,
	packetid.ServerboundBlockEntityTagQuery:      nil,
	packetid.ServerboundChangeDifficulty:         nil,
	packetid.ServerboundChatCommand:              nil,
	packetid.ServerboundChat:                     nil,
	packetid.ServerboundChatPreview:              nil,
	packetid.ServerboundClientCommand:            nil,
	packetid.ServerboundClientInformation:        clientInformation{},
	packetid.ServerboundCommandSuggestion:        nil,
	packetid.ServerboundContainerButtonClick:     nil,
	packetid.ServerboundContainerClick:           nil,
	packetid.ServerboundContainerClose:           nil,
	packetid.ServerboundCustomPayload:            nil,
	packetid.ServerboundEditBook:                 nil,
	packetid.ServerboundEntityTagQuery:           nil,
	packetid.ServerboundInteract:                 nil,
	packetid.ServerboundJigsawGenerate:           nil,
	packetid.ServerboundKeepAlive:                nil,
	packetid.ServerboundLockDifficulty:           nil,
	packetid.ServerboundMovePlayerPos:            clientMovePlayerPos{},
	packetid.ServerboundMovePlayerPosRot:         clientMovePlayerPosRot{},
	packetid.ServerboundMovePlayerRot:            clientMovePlayerRot{},
	packetid.ServerboundMovePlayerStatusOnly:     clientMovePlayerStatusOnly{},
	packetid.ServerboundMoveVehicle:              clientMoveVehicle{},
	packetid.ServerboundPaddleBoat:               nil,
	packetid.ServerboundPickItem:                 nil,
	packetid.ServerboundPlaceRecipe:              nil,
	packetid.ServerboundPlayerAbilities:          nil,
	packetid.ServerboundPlayerAction:             nil,
	packetid.ServerboundPlayerCommand:            nil,
	packetid.ServerboundPlayerInput:              nil,
	packetid.ServerboundPong:                     nil,
	packetid.ServerboundRecipeBookChangeSettings: nil,
	packetid.ServerboundRecipeBookSeenRecipe:     nil,
	packetid.ServerboundRenameItem:               nil,
	packetid.ServerboundResourcePack:             nil,
	packetid.ServerboundSeenAdvancements:         nil,
	packetid.ServerboundSelectTrade:              nil,
	packetid.ServerboundSetBeacon:                nil,
	packetid.ServerboundSetCarriedItem:           nil,
	packetid.ServerboundSetCommandBlock:          nil,
	packetid.ServerboundSetCommandMinecart:       nil,
	packetid.ServerboundSetCreativeModeSlot:      nil,
	packetid.ServerboundSetJigsawBlock:           nil,
	packetid.ServerboundSetStructureBlock:        nil,
	packetid.ServerboundSignUpdate:               nil,
	packetid.ServerboundSwing:                    nil,
	packetid.ServerboundTeleportToEntity:         nil,
	packetid.ServerboundUseItemOn:                nil,
	packetid.ServerboundUseItem:                  nil,
}
