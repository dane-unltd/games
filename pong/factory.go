package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	"github.com/dane-unltd/linalg/matrix"
)

func createPlayer(id core.EntId, playerNo helpers.Uint32, mut core.MutFuncs) {
	mut.Mutate("players", id, playerNo)
	mut.Mutate("score", id, helpers.Uint32(3))

	pos := matrix.VecD{(float64(playerNo) - 1.5) * 150, 0, 0}
	size := matrix.VecD{5, 20, 5}
	ori := matrix.VecD{1, 0, 0}
	mut.Mutate("pos", id, pos)
	mut.Mutate("vel", id, matrix.NewVecD(3))
	mut.Mutate("size", id, size)
	mut.Mutate("ori", id, ori)
	mut.Mutate("massinv", id, helpers.Float64(0))
}

func createBall(mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("ball", id, core.Empty{})

	vel := matrix.VecD{-3, 0, 0}
	size := matrix.VecD{5, 5, 5}
	ori := matrix.VecD{1, 0, 0}
	mut.Mutate("pos", id, matrix.NewVecD(3))
	mut.Mutate("vel", id, vel)
	mut.Mutate("size", id, size)
	mut.Mutate("ori", id, ori)
	mut.Mutate("massinv", id, helpers.Float64(1))
}
