package main

import (
	"github.com/dane-unltd/core"
)

func createPlayer(id core.EntId, playerNo core.Uint32, mut core.MutFuncs) {
	mut.Mutate("players", id, playerNo)

	pos := core.Vec3{(float64(playerNo) - 1.5) * 100, 0, 0}
	size := core.Vec3{5, 20, 5}
	ori := core.Vec3{1, 0, 0}
	mut.Mutate("pos", id, pos)
	mut.Mutate("vel", id, core.Zeros())
	mut.Mutate("size", id, size)
	mut.Mutate("ori", id, ori)
}

func createBall(mut core.MutFuncs) {
	id := mut.NewId()
	mut.Mutate("ball", id, core.Empty{})

	vel := core.Vec3{0, 10, 0}
	size := core.Vec3{5, 5, 5}
	ori := core.Vec3{1, 0, 0}
	mut.Mutate("pos", id, core.Zeros())
	mut.Mutate("vel", id, vel)
	mut.Mutate("size", id, size)
	mut.Mutate("ori", id, ori)
}
