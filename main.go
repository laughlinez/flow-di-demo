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


	go func() {
		c := flow.NewCircuit()

		c.Add("xx", "XXX")
		//c.Add("db", "LevelDB")

		c.Add("forever", "Forever")
		c.Run()
	}()


	c1 := flow.NewCircuit()

	c1.Add("yy", "YYY")

	c1.Add("forever", "Forever")
	c1.Run()






}
