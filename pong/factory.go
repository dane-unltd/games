package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	. "github.com/dane-unltd/linalg/matrix"
)

func createPlayer(id core.EntId, playerNo helpers.Uint32, mut core.MutFuncs) {
	mut.Mutate("players", id, playerNo)
	mut.Mutate("score", id, helpers.Uint32(3))

	pos := VecD{(float64(playerNo) - 1.5) * 150, 0, 0}
	scale := VecD{5, 10, 5}
	rot := Eye(3)
	mut.Mutate("pos", id, pos)
	mut.Mutate("vel", id, NewVecD(3))
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(0.1))
	mut.Mutate("model", id, helpers.Uint32(1))
}

func createWall(pos, scale VecD, rot *DenseD, mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("pos", id, pos)
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(0))
	mut.Mutate("vel", id, NewVecD(3))
	mut.Mutate("model", id, helpers.Uint32(1))
}

func createBall(mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("ball", id, core.Empty{})

	vel := VecD{-3, 0, 0}
	scale := VecD{15, 15, 15}
	rot := Eye(3)
	mut.Mutate("pos", id, NewVecD(3))
	mut.Mutate("vel", id, vel)
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(1))
	mut.Mutate("model", id, helpers.Uint32(2))
}
