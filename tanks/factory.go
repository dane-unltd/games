package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	"github.com/dane-unltd/engine/physics"
	. "github.com/dane-unltd/linalg/matrix"
)

func createPlayer(state core.StateMap, id core.EntId, playerNo helpers.Uint32, mut core.MutFuncs) {
	mut.Mutate("players", id, playerNo)
	mut.Mutate("score", id, helpers.Uint32(3))

	pos := VecD{(float64(playerNo) - 1.5) * 150, 0, 0}
	scale := VecD{15, 15, 5}
	rot := VecD{1, 0, 0, 0}
	mut.Mutate("pos", id, pos)
	mut.Mutate("vel", id, NewVecD(3))
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(0.1))
	mut.Mutate("model", id, helpers.Uint32(1))
	mut.Mutate("supportfunc", id, physics.LinOptPoly(physics.AABB(VecD{1, 1, 1})))
	mut.Mutate("newrigbod", id, core.Empty{})

}

func createWall(state core.StateMap, pos, scale, rot VecD, mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("pos", id, pos)
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(0))
	mut.Mutate("vel", id, NewVecD(3))
	mut.Mutate("model", id, helpers.Uint32(3))
	mut.Mutate("supportfunc", id, physics.LinOptPoly(physics.AABB(VecD{1, 1, 1})))
	mut.Mutate("newrigbod", id, core.Empty{})
}

func createBall(state core.StateMap, mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("ball", id, core.Empty{})

	vel := VecD{-3, 0, 0}
	scale := VecD{15, 15, 15}
	rot := VecD{1, 0, 0, 0}
	mut.Mutate("pos", id, NewVecD(3))
	mut.Mutate("vel", id, vel)
	mut.Mutate("scale", id, scale)
	mut.Mutate("rot", id, rot)
	mut.Mutate("massinv", id, helpers.Float64(1))
	mut.Mutate("model", id, helpers.Uint32(2))
	mut.Mutate("supportfunc", id, physics.LinOptSphere(1))
	mut.Mutate("newrigbod", id, core.Empty{})
}
