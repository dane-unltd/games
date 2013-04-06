package main

import (
	"github.com/dane-unltd/core"
	"log"
	"os"
)

const (
	cmdUp   = 0
	cmdDown = 1
)

func move(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	pos := state["pos"].(core.VecMap)
	vel := state["vel"].(core.VecMap)

	for i := range vel {
		if vel[i] != core.Zeros() {
			newPos := core.Add(pos[i], vel[i])
			newVel := vel[i]
			if newPos[1] > 50 {
				newPos[1] = 50
				if newVel[1] > 0 {
					newVel[1] = -newVel[1]
				}
			} else if newPos[1] < -50 {
				newPos[1] = -50
				if newVel[1] < 0 {
					newVel[1] = -newVel[1]
				}
			}
			mut.Mutate("pos", i, newPos)
			mut.Mutate("vel", i, newVel)
		}
	}
}

func processInput(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	input := state["input"].(core.IdMap)
	players := state["players"].(core.IdMap)

	for id, _ := range players {
		cmd := input[id].(core.UserCmd)
		v := core.Zeros()
		if cmd.Active(cmdUp) {
			v[1] += 10
		}
		if cmd.Active(cmdDown) {
			v[1] -= 10
		}
		mut.Mutate("vel", id, v)
	}
}

func handleLogin(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	logins := state["logins"].(core.IdMap)
	players := state["players"].(core.IdMap)

	if len(players) == 2 {
		return
	}

	if len(players) == 1 {
		pl := core.Uint32(0)
		for _, p := range players {
			pl = p.(core.Uint32)
		}
		pl2 := core.Uint32(0)
		if pl == 1 {
			pl2 = 2
		} else {
			pl2 = 1
		}
		for id, _ := range logins {
			createPlayer(id, pl2, mut)
			createBall(mut)
			return
		}
	}
	i := 0
	for id, _ := range logins {
		i++
		createPlayer(id, core.Uint32(i), mut)
		if i == 2 {
			createBall(mut)
			return
		}
	}
}

func destroyPlayer(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	discs := state["disconnects"].(core.IdList)
	players := state["players"].(core.IdMap)
	ball := state["ball"].(core.IdList)

	for id := range discs {
		log.Println("destroying ", id)
		_, ok := players[id]

		if ok {
			for ballId := range ball {
				mut.Destroy(ballId)
			}
		}

		mut.Destroy(id)
	}
}

func resolveCollisions(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	pos := state["pos"].(core.VecMap)
	vel := state["vel"].(core.VecMap)
	size := state["size"].(core.VecMap)
	massInv := state["massinv"].(core.IdMap)

	ids := make([]EntId, len(pos))
	i := 0
	for id := range pos {
		ids[i] = id
		i++
	}

	for ix1 := 0; ix1 < len(ids); ix1++ {
		for ix2 := ix1 + 1; ix2 < len(ids); ix2++ {
			p1 := pos[ids[ix1]]
			p2 := pos[ids[ix2]]
			v1 := vel[ids[ix1]]
			v2 := vel[ids[ix2]]
			s1 := size[ids[ix1]]
			s2 := size[ids[ix2]]
			mi1 := massInv[ids[ix1]].(core.Float64)
			mi2 := massInv[ids[ix2]].(core.Float64)

			dp := core.Sub(p1, p2)
			dv := core.Sub(v1, v2)
			ds := core.Mul(core.Add(s1, s2), 0.5)

			dist := core.MulH(core.Sign(dp),
				core.Max(core.Sub(core.Abs(dp), ds), 0))
			n := core.Neg(core.Normalize(dist))

			nV := core.Dot(dv, n)
			remove := nV + core.Lenth(dist)
			if remove < 0 {
				imp := remove / (mi1 + mi2)
				v1 = core.Add(v1, core.Mul(n, imp))
				v2 = core.Sub(v2, core.Mul(n, imp))
			}
		}
	}
}

func entSel(id core.EntId, state core.StateMap) []core.EntId {
	pos := state["pos"].(core.VecMap)
	list := make([]core.EntId, 0, 10)
	for id := range pos {
		list = append(list, id)
	}
	return list
}

func main() {
	path := os.Getenv("HOME") + "/nginx/ws/pong"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	srv := core.NewWsServer(path, 30)

	info := core.SerInfo{entSel, make([]string, 4)}
	info.States[0] = "pos"
	info.States[1] = "vel"
	info.States[2] = "size"
	info.States[3] = "ori"

	srv.SetSerInfo(info)

	srv.AddState("pos", core.NewVecMap())
	srv.AddState("vel", core.NewVecMap())
	srv.AddState("size", core.NewVecMap())
	srv.AddState("ori", core.NewVecMap())
	srv.AddState("players", core.NewIdMap())
	srv.AddState("ball", core.NewIdList())

	srv.AddTransFunc(0, handleLogin)
	srv.AddTransFunc(0, destroyPlayer)
	srv.AddTransFunc(1, processInput)
	srv.AddTransFunc(2, move)

	srv.Run()

	log.Println("done")
}
