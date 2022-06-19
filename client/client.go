package client

import (
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/net"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/go-mc/server/world"
	"go.uber.org/zap"
)

type Client struct {
	log         *zap.Logger
	conn        *net.Conn
	handlers    []packetHandler
	player      Player
	chunkLoader *world.Loader
}

func (c *Client) ViewChunkLoad(pos level.ChunkPos, chunk *level.Chunk) {
	c.SendLevelChunkWithLight(pos, chunk)
}

func (c *Client) ViewChunkUnload(pos level.ChunkPos) {
	//TODO implement me
	panic("implement me")
}

type packetHandler interface {
	Handle(p pk.Packet, c *Client) error
}

func New(log *zap.Logger, conn *net.Conn) *Client {
	return &Client{
		log:      log,
		conn:     conn,
		handlers: defaultHandlers,
	}
}

func (c *Client) Spawn(p Player, w *world.World) error {
	c.player = p
	p.SetWorld(w)
	c.chunkLoader = world.NewLoader(w, p.ChunkPos(), p.ChunkRadius())
	w.AddLoader(c.chunkLoader, c)
	err := c.SendLogin(p)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Start() {
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
		err = c.handlers[packet.ID].Handle(packet, c)
		if err != nil {
			c.log.Error("Handle packet error", zap.Int32("id", packet.ID), zap.Error(err))
			return
		}
	}
}

var defaultHandlers = []packetHandler{
	packetid.ServerboundAcceptTeleportation:      nil,
	packetid.ServerboundBlockEntityTagQuery:      nil,
	packetid.ServerboundChangeDifficulty:         nil,
	packetid.ServerboundChatCommand:              nil,
	packetid.ServerboundChat:                     nil,
	packetid.ServerboundChatPreview:              nil,
	packetid.ServerboundClientCommand:            nil,
	packetid.ServerboundClientInformation:        clientInfomation{},
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
	packetid.ServerboundMovePlayerPos:            nil,
	packetid.ServerboundMovePlayerPosRot:         nil,
	packetid.ServerboundMovePlayerRot:            nil,
	packetid.ServerboundMovePlayerStatusOnly:     nil,
	packetid.ServerboundMoveVehicle:              nil,
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
