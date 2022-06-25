package world

import (
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/level"
)

type Client interface {
	ChunkViewer
	EntityViewer
	SendDisconnect(reason chat.Message)
	SendPlayerPosition(pos [3]float64, rot [2]float32, dismountVehicle bool) (teleportID int32)
}

type ChunkViewer interface {
	ViewChunkLoad(pos level.ChunkPos, c *level.Chunk)
	ViewChunkUnload(pos level.ChunkPos)
}

type EntityViewer interface {
	ViewAddPlayer(p *Player)
	ViewRemoveEntities(entityIDs []int32)
	ViewMoveEntityPos(id int32, delta [3]int16, onGround bool)
	ViewMoveEntityPosAndRot(id int32, delta [3]int16, rot [2]int8, onGround bool)
	ViewMoveEntityRot(id int32, rot [2]int8, onGround bool)
	ViewRotateHead(id int32, yaw int8)
	ViewTeleportEntity(id int32, pos [3]float64, rot [2]float32, onGround bool)
}
