package main

import (
	"github.com/laughlinez/flow"
	demogadgets  "github.com/laughlinez/flow-di-demo/gadgets"
	_  "github.com/laughlinez/flow-di-demo/gadgets/database"
)


func main () {

	flow.Config["DATA_DIR"] = "./data"

	x := new(demogadgets.Foo)
	y := new(demogadgets.Bar)


	flow.Registry["XXX"] = func() flow.Circuitry { return x }
	flow.Registry["YYY"] = func() flow.Circuitry { return y }

	c := flow.NewCircuit()

	c.Add("xx", "XXX")
	c.Add("yy", "YYY")
	c.Add("db", "LevelDB")

	c.Add("forever", "Forever")
	c.Run()





}
