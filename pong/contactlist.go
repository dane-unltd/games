package main

import (
	"github.com/dane-unltd/engine/core"
	"github.com/dane-unltd/engine/physics"
	"io"
)

type IdPair struct {
	a, b core.EntId
}

type ContactList map[IdPair]*physics.Contact

func NewContactList() ContactList {
	return make(ContactList)
}

func (c ContactList) Clone() core.State {
	newCL := make(ContactList)
	for k, v := range c {
		newCL[k] = v.Copy().(*physics.Contact)
	}
	return newCL
}
func (c ContactList) Mutate(id core.EntId, value interface{}) {
	if value == nil {
		remove := make([]IdPair, 0)
		for ids := range c {
			if ids.a == id || ids.b == id {
				remove = append(remove, ids)
			}
		}
		for _, ids := range remove {
			delete(c, ids)
		}
		return
	}
	con := value.(*physics.Contact)
	a := con.A
	b := con.B
	p := IdPair{}
	if a < b {
		p.a, p.b = a, b
	} else {
		p.a, p.b = b, a
	}
	c[p] = con
}
func (c ContactList) SerDiff(buf io.Writer, newEnts []core.EntId, newSt core.State) {
	//TODO
}
func (c ContactList) DeserDiff(buf io.Reader, newEnts []core.EntId) {
	//TODO
}
