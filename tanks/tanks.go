package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/helpers"
	"github.com/dane-unltd/engine/physics"
	_ "github.com/dane-unltd/linalg/blas"
	. "github.com/dane-unltd/linalg/matrix"
	"log"
	"math"
	"os"
)

const (
	cmdUp    = 0
	cmdDown  = 1
	cmdLeft  = 2
	cmdRight = 3
)

func initial(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	if time != 1 {
		return
	}
	//walls
	pos := VecD{0, 60, 0}
	size := VecD{200, 5, 50}
	rot := VecD{1, 0, 0, 0}
	createWall(state, pos, size, rot, mut)

	pos = VecD{0, -60, 0}
	size = VecD{200, 5, 50}
	rot = VecD{1, 0, 0, 0}
	createWall(state, pos, size, rot, mut)

	pos = VecD{100, 0, 0}
	size = VecD{5, 120, 50}
	rot = VecD{1, 0, 0, 0}
	createWall(state, pos, size, rot, mut)

	pos = VecD{-100, 0, 0}
	size = VecD{5, 120, 50}
	rot = VecD{1, 0, 0, 0}
	createWall(state, pos, size, rot, mut)

	//floor
	pos = VecD{0, 0, -30}
	size = VecD{200, 150, 10}
	rot = VecD{1, 0, 0, 0}
	createWall(state, pos, size, rot, mut)
}

func move(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	pos := state["pos"].(helpers.VecMap)
	vel := state["vel"].(helpers.VecMap)

	for id := range vel {
		if !vel[id].Equals(NewVecD(3)) {
			newVel := vel[id]
			newVel[2] = 0
			newPos := NewVecD(3).Add(pos[id], newVel)
			mut.Mutate("pos", id, newPos)
			mut.Mutate("vel", id, newVel)
		}
	}
}

func processInput(time core.Tick, state core.StateMap, mut core.MutFuncs) {
	input := state["input"].(core.IdMap)
	players := state["players"].(core.IdMap)
	rot := state["rot"].(helpers.VecMap)

	for id, _ := range players {
		rotOld := physics.Quaternion(rot[id])
		cmd := input[id].(core.UserCmd)
		v := NewVecD(3)
		rotAdd := physics.Quaternion(NewVecD(4))
		if cmd.Active(cmdUp) {
			v[1] += 3
		}
		if cmd.Active(cmdDown) {
			v[1] -= 3
		}
		if cmd.Active(cmdLeft) {
			rotAdd.Normalize(physics.Quaternion{10, 0, 0, 1})
			rotOld.Mul(rotAdd, rotOld)
		}
		if cmd.Active(cmdRight) {
			rotAdd.Normalize(physics.Quaternion{10, 0, 0, -1})
			rotOld.Mul(rotAdd, rotOld)
		}
		rotOld.Normalize(rotOld)
		R := physics.RotFromQuat(rotOld)
		log.Println(rotOld, R)
		vNew := NewVecD(3)
		vNew.Mul(R, v)
		mut.Mutate("vel", id, vNew)
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
	rot := state["rot"].(helpers.VecMap)
	supp := state["supportfunc"].(core.IdMap)
	massInv := state["massinv"].(core.IdMap)
	contacts := state["contact"].(ContactList)

	newVel := vel.Clone().(helpers.VecMap)

	for iter := 0; iter < 10; iter++ {
		for ids, c := range contacts {
			pA, pB := pos[ids.a], pos[ids.b]
			vA, vB := newVel[ids.a], newVel[ids.b]
			rotA := physics.RotFromQuat(physics.Quaternion(rot[ids.a]))
			rotB := physics.RotFromQuat(physics.Quaternion(rot[ids.b]))
			scaleA, scaleB := scale[ids.a], scale[ids.b]
			suppA := supp[ids.a].(physics.SupportFunc)
			suppB := supp[ids.b].(physics.SupportFunc)

			transA := NewDenseD(3, 3)
			transA.Mul(rotA, DiagD(scaleA))
			transB := NewDenseD(3, 3)
			transB.Mul(rotB, DiagD(scaleB))

			c.Update(pA, pB, transA, transB, suppA, suppB)

			dv := NewVecD(3).Sub(vA, vB)

			nV := Ddot(c.Normal, dv)

			remove := c.Dist - 1 - nV

			if math.IsNaN(remove) {
				log.Println(remove, c.Dist, nV)
				panic("argh")
			}
			if remove < 0 {
				log.Println(remove)
				minvA := float64(massInv[ids.a].(helpers.Float64))
				minvB := float64(massInv[ids.b].(helpers.Float64))
				imp := remove / (minvA + minvB)
				if math.IsNaN(imp) {
					log.Println(remove, imp, minvA, minvB, c.Dist, nV)
					panic("argh")
				}

				newVel[ids.a].Add(newVel[ids.a],
					NewVecD(3).Axpy(imp*minvA, c.Normal))
				newVel[ids.b].Sub(newVel[ids.b],
					NewVecD(3).Axpy(imp*minvB, c.Normal))
				if math.IsNaN(newVel[ids.a][0]) {
					log.Println(c.Normal, remove, imp, minvA, minvB, c.Dist, nV)
					panic("argh")
				}
			}
			mut.Mutate("contact", 0, c)
		}
	}
	for id, vel := range newVel {
		mut.Mutate("vel", id, vel)
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
			log.Println("new Contact ", id1, id2)
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
	path := os.Getenv("HOME") + "/nginx/ws/tanks"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	srv := core.NewWsServer(path, 30)

	info := core.SerInfo{entSel, make([]string, 6)}
	info.States[0] = "pos"
	info.States[1] = "vel"
	info.States[2] = "scale"
	info.States[3] = "rot"
	info.States[4] = "model"
	info.States[5] = "score"

	srv.SetSerInfo(info)

	srv.AddState("pos", helpers.NewVecMap())
	srv.AddState("vel", helpers.NewVecMap())
	srv.AddState("scale", helpers.NewVecMap())
	srv.AddState("rot", helpers.NewVecMap())
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

	srv.Run()

	log.Println("done")
}
