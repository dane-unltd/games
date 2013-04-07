package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	"github.com/dane-unltd/engine/physics"
	_ "github.com/dane-unltd/linalg/blasinit"
	_ "github.com/dane-unltd/linalg/lapackinit"
	"github.com/dane-unltd/linalg/matrix"
	"log"
	"math"
	"os"
)

const (
	cmdUp   = 0
	cmdDown = 1
)

func move(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	pos := state["pos"].(helpers.VecMap)
	vel := state["vel"].(helpers.VecMap)

	for i := range vel {
		if !vel[i].Equals(matrix.ZeroVec(3)) {
			newPos := matrix.ZeroVec(3).Add(pos[i], vel[i])
			newVel := vel[i]
			newVel[2] = 0
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
			if newPos[0] > 100 {
				newPos[0] = 100
				if newVel[0] > 0 {
					newVel[0] = -newVel[0]
				}
			} else if newPos[0] < -100 {
				newPos[0] = -100
				if newVel[0] < 0 {
					newVel[0] = -newVel[0]
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
		v := matrix.ZeroVec(3)
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
		pl := helpers.Uint32(0)
		for _, p := range players {
			pl = p.(helpers.Uint32)
		}
		pl2 := helpers.Uint32(0)
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
		createPlayer(id, helpers.Uint32(i), mut)
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
	pos := state["pos"].(helpers.VecMap)
	vel := state["vel"].(helpers.VecMap)
	size := state["size"].(helpers.VecMap)
	massInv := state["massinv"].(core.IdMap)

	ids := make([]core.EntId, len(pos))
	rb := make([]*physics.RigidBody, len(pos))
	i := 0
	for id := range pos {
		ids[i] = id
		v := size[id].Copy()
		v.Mul(0.5, v)
		A := matrix.FromArrayD(v, false, 3, 1)
		for ix := 0; ix < 3; ix++ {
			v[ix] = -v[ix]
			A.AddCol(v)
			v[ix] = -v[ix]
		}
		for ix := 0; ix < 3; ix++ {
			v[ix] = -v[ix]
		}
		A.AddCol(v)
		for ix := 0; ix < 3; ix++ {
			v[ix] = -v[ix]
			A.AddCol(v)
			v[ix] = -v[ix]
		}

		rb[i] = new(physics.RigidBody)
		rb[i].Pos = pos[id]
		rb[i].Rot = matrix.Eye(3)
		rb[i].MassInv = float64(massInv[id].(helpers.Float64))
		rb[i].Points = A

		i++
	}

	for ix1 := 0; ix1 < len(ids); ix1++ {
		for ix2 := ix1 + 1; ix2 < len(ids); ix2++ {
			c := physics.CreateContact(rb[ix1], rb[ix2])

			v1 := vel[ids[ix1]]
			v2 := vel[ids[ix2]]

			dv := matrix.ZeroVec(3).Sub(v1, v2)
			dp := matrix.ZeroVec(3).Sub(rb[ix2].Pos, rb[ix1].Pos)

			vProj := matrix.ZeroVec(3)
			vProj.Mul(c.Normal.Dot(dv), c.Normal)
			nV := 0.0
			if dp.Dot(vProj) > 0 {
				nV = math.Abs(c.Normal.Dot(dv))
			}
			remove := c.Dist - nV

			if remove < 0 {
				log.Println("Normal", c.Normal)
				log.Println("vproj", vProj)
				log.Println("dp", dp)
				log.Println("c.Dist", c.Dist)
				vProj.Normalize(vProj)
				imp := -nV / (c.A.MassInv + c.B.MassInv)
				v1New := matrix.ZeroVec(3).Add(v1,
					matrix.ZeroVec(3).Mul(2*imp*c.A.MassInv, vProj))
				v2New := matrix.ZeroVec(3).Sub(v2,
					matrix.ZeroVec(3).Mul(2*imp*c.B.MassInv, vProj))
				mut.Mutate("vel", ids[ix1], v1New)
				mut.Mutate("vel", ids[ix2], v2New)
				log.Println("newVels", v1New, v2New)
			}
		}
	}
}

func checkBall(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	ball := state["ball"].(core.IdList)
	score := state["score"].(core.IdMap)
	pos := state["pos"].(helpers.VecMap)
	players := state["players"].(core.IdMap)

	var p1, p2 core.EntId
	for id, v := range players {
		if v.(helpers.Uint32) == 1 {
			p1 = id
		}
		if v.(helpers.Uint32) == 2 {
			p2 = id
		}
	}

	var ballId core.EntId
	for id := range ball {
		ballId = id
	}

	if (p1 == 0) || (p2 == 0) || (ballId == 0) {
		return
	}

	ballPos := pos[ballId]
	if ballPos[0] < -70 {
		vel := matrix.VecD{-3, 0, 0}
		mut.Mutate("pos", ballId, matrix.ZeroVec(3))
		mut.Mutate("vel", ballId, vel)
		mut.Mutate("score", p2, score[p2].(helpers.Uint32)+1)
		log.Println(score)
	}
	if ballPos[0] > 70 {
		vel := matrix.VecD{3, 0, 0}
		mut.Mutate("pos", ballId, matrix.ZeroVec(3))
		mut.Mutate("vel", ballId, vel)
		mut.Mutate("score", p1, score[p1].(helpers.Uint32)+1)
		log.Println(score)
	}

}

func entSel(id core.EntId, state core.StateMap) []core.EntId {
	pos := state["pos"].(helpers.VecMap)
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

	info := core.SerInfo{entSel, make([]string, 5)}
	info.States[0] = "pos"
	info.States[1] = "vel"
	info.States[2] = "size"
	info.States[3] = "ori"
	info.States[4] = "score"

	srv.SetSerInfo(info)

	srv.AddState("pos", helpers.NewVecMap())
	srv.AddState("vel", helpers.NewVecMap())
	srv.AddState("size", helpers.NewVecMap())
	srv.AddState("ori", helpers.NewVecMap())
	srv.AddState("players", core.NewIdMap())
	srv.AddState("score", core.NewIdMap())
	srv.AddState("massinv", core.NewIdMap())
	srv.AddState("ball", core.NewIdList())

	srv.AddTransFunc(0, handleLogin)
	srv.AddTransFunc(0, destroyPlayer)
	srv.AddTransFunc(1, processInput)
	srv.AddTransFunc(2, resolveCollisions)
	srv.AddTransFunc(3, move)
	srv.AddTransFunc(4, checkBall)

	srv.Run()

	log.Println("done")
}
