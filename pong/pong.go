package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	"github.com/dane-unltd/engine/physics"
	_ "github.com/dane-unltd/linalg/blas"
	. "github.com/dane-unltd/linalg/matrix"
	"log"
	"math"
	"math/rand"
	"os"
)

const (
	cmdUp   = 0
	cmdDown = 1
)

func initial(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	if time != 1 {
		return
	}
	pos := VecD{0, 60, 0}
	size := VecD{200, 5, 30}
	rot := Eye(3)
	createWall(state, pos, size, rot, mut)

	pos = VecD{0, -60, 0}
	size = VecD{200, 5, 30}
	rot = Eye(3)
	createWall(state, pos, size, rot, mut)
}

func move(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	pos := state["pos"].(helpers.VecMap)
	vel := state["vel"].(helpers.VecMap)
	players := state["players"].(core.IdMap)

	for id := range vel {
		if !vel[id].Equals(NewVecD(3)) {
			newVel := vel[id]
			newVel[2] = 0
			if _, ok := players[id]; ok {
				newVel[0] = 0
			}
			newPos := NewVecD(3).Add(pos[id], newVel)
			mut.Mutate("pos", id, newPos)
			mut.Mutate("vel", id, newVel)
		}
	}
}

func processInput(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	input := state["input"].(core.IdMap)
	players := state["players"].(core.IdMap)

	for id, _ := range players {
		cmd := input[id].(core.UserCmd)
		v := NewVecD(3)
		if cmd.Active(cmdUp) {
			v[1] += 3
		}
		if cmd.Active(cmdDown) {
			v[1] -= 3
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
			createPlayer(state, id, pl2, mut)
			createBall(state, mut)
			return
		}
	}
	i := 0
	for id, _ := range logins {
		i++
		createPlayer(state, id, helpers.Uint32(i), mut)
		if i == 2 {
			createBall(state, mut)
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
	scale := state["scale"].(helpers.VecMap)
	rot := state["rot"].(core.IdMap)
	supp := state["supportfunc"].(core.IdMap)
	massInv := state["massinv"].(core.IdMap)
	contacts := state["contact"].(ContactList)

	for ids, c := range contacts {
		pA, pB := pos[ids.a], pos[ids.b]
		vA, vB := vel[ids.a], vel[ids.b]
		rotA, rotB := rot[ids.a].(*DenseD), rot[ids.b].(*DenseD)
		scaleA, scaleB := scale[ids.a], scale[ids.b]
		suppA := supp[ids.a].(physics.SupportFunc)
		suppB := supp[ids.b].(physics.SupportFunc)

		transA := NewDenseD(3, 3)
		transA.Mul(rotA, DiagD(scaleA))
		transB := NewDenseD(3, 3)
		transB.Mul(rotB, DiagD(scaleB))

		c.Update(pA, pB, transA, transB, suppA, suppB)

		dv := NewVecD(3).Sub(vA, vB)
		dp := NewVecD(3).Sub(pB, pA)

		vProj := NewVecD(3)
		vProj.Axpy(Ddot(c.Normal, dv), c.Normal)
		nV := 0.0
		if Ddot(dp, vProj) > 0 {
			nV = math.Abs(Ddot(c.Normal, dv))
		}
		remove := c.Dist - 0.1 - nV

		if remove < 0 {
			minvA := float64(massInv[ids.a].(helpers.Float64))
			minvB := float64(massInv[ids.b].(helpers.Float64))
			vProj.Normalize(vProj)
			imp := -nV / (minvA + minvB)
			vANew := NewVecD(3).Add(vA,
				NewVecD(3).Axpy(2*imp*minvA, vProj))
			vBNew := NewVecD(3).Sub(vB,
				NewVecD(3).Axpy(2*imp*minvB, vProj))
			mut.Mutate("vel", ids.a, vANew)
			mut.Mutate("vel", ids.b, vBNew)
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
	if ballPos[0] < -110 {
		vely := (rand.Float64() - 0.5) * 4
		vel := VecD{-3, vely, 0}
		mut.Mutate("pos", ballId, NewVecD(3))
		mut.Mutate("vel", ballId, vel)
		mut.Mutate("score", p2, score[p2].(helpers.Uint32)+1)
		log.Println(score)
	}
	if ballPos[0] > 110 {
		vely := (rand.Float64() - 0.5) * 4
		vel := VecD{3, vely, 0}
		mut.Mutate("pos", ballId, NewVecD(3))
		mut.Mutate("vel", ballId, vel)
		mut.Mutate("score", p1, score[p1].(helpers.Uint32)+1)
		log.Println(score)
	}

}

func createContacts(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	massinv := state["massinv"].(core.IdMap)
	newrb := state["newrigbod"].(core.IdList)

	for id1 := range newrb {
		mut.Mutate("newrigbod", id1, nil)
		minv1 := float64(massinv[id1].(helpers.Float64))
		exclude := minv1 == 0
		for id2, minv2 := range massinv {
			if id1 == id2 {
				continue
			}
			if minv2.(helpers.Float64) == 0 && exclude {
				continue
			}
			c := physics.NewContact()
			if id1 < id2 {
				c.A = id1
				c.B = id2
			} else {
				c.A = id2
				c.B = id1
			}
			mut.Mutate("contact", 0, c)
		}
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
	info.States[2] = "scale"
	info.States[3] = "model"
	info.States[4] = "score"

	srv.SetSerInfo(info)

	srv.AddState("pos", helpers.NewVecMap())
	srv.AddState("vel", helpers.NewVecMap())
	srv.AddState("scale", helpers.NewVecMap())
	srv.AddState("rot", core.NewIdMap())
	srv.AddState("supportfunc", core.NewIdMap())
	srv.AddState("contact", NewContactList())
	srv.AddState("newrigbod", core.NewIdList())

	srv.AddState("players", core.NewIdMap())
	srv.AddState("model", core.NewIdMap())
	srv.AddState("score", core.NewIdMap())
	srv.AddState("massinv", core.NewIdMap())
	srv.AddState("ball", core.NewIdList())

	srv.AddTransFunc(0, initial)
	srv.AddTransFunc(0, handleLogin)
	srv.AddTransFunc(0, createContacts)
	srv.AddTransFunc(0, destroyPlayer)
	srv.AddTransFunc(1, processInput)
	srv.AddTransFunc(2, resolveCollisions)
	srv.AddTransFunc(3, move)
	srv.AddTransFunc(4, checkBall)

	srv.Run()

	log.Println("done")
}
