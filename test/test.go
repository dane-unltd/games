package main

import (
	"fmt"
	"github.com/dane-unltd/core"
	"os"
	"time"
)

func move(time uint64, input core.CmdSrc, state core.StateMap,
	mut core.MutFuncs) {
	pos := state["pos"].(core.VecMap)
	vel := state["vel"].(core.VecMap)

	for i := range vel {
		if vel[i] != 0.0 {
			newPos := pos[i] + vel[i]
			mut.Mutate("pos", i, newPos)
		}
	}
}

func createPlayer(mut core.MutFuncs, id core.EntId) {
	mut.Mutate("pos", id, core.Vec3(0))
	mut.Mutate("vel", id, core.Vec3(1))
}

func entSel(id core.EntId, state core.StateMap) core.IdList {
	pos := state["pos"].(core.VecMap)
	list := make(core.IdList, 0, 10)
	for id := range pos {
		list = append(list, id)
	}
	return list
}

func main() {
	pos := core.NewVecMap()
	vel := core.NewVecMap()

	_ = pos
	_ = vel

	srv := core.NewServer(":33333")
	cl := core.NewClient("blubb", os.Args[1], "localhost:33333")

	info := make([]core.SerInfo, 2)
	info[0] = core.SerInfo{"pos"}
	info[1] = core.SerInfo{"vel"}

	srv.SetSerFuncs(core.SerFuncFact(info, entSel))
	cl.SetSerFuncs(core.SerFuncFact(info, entSel))

	srv.AddState("pos", pos)
	srv.AddState("vel", vel)

	cl.AddState("pos", pos)
	cl.AddState("vel", vel)

	srv.AddTransFunc(0, move)

	cl.AddTransFunc(0, move)

	srv.AddFactFunc("player", createPlayer)

	go srv.Run()

	time.Sleep(1e8)

	cl.Run()

	time.Sleep(1e8)

	fmt.Println("done")
}
