package client

import (
	"github.com/Tnze/go-mc/data/packetid"
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

func (c *Client) JoinWorld(p Player, w *world.World) {
	c.player = p
	//c.chunkLoader
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
