package demogadget

import (
	"github.com/laughlinez/flow"
	flowapi "github.com/laughlinez/flow/api"
	"time"
	"fmt"
)

func init() {
	fmt.Printf("Demo Gadgets Init...")
}

type Foo struct {
	flow.Gadget
	z int
	A int
	In flow.Input
	Out flow.Output
	Settings flowapi.ISettingsAPI `gadget:"SettingsAPI"`  //gadget requests ISettingsAPI
	DB flowapi.IDBReadWriteAPI `gadget:"DBReadWriteAPI"`  //gadget requests IDBReadWriteAPI
}


type Bar struct {
	flow.Gadget
	In flow.Input
	Out flow.Output
	Settings flowapi.ISettingsAPI `gadget:"SettingsAPI"` //gadget requests ISettingsAPI
	DB flowapi.IDBReadWriteAPI `gadget:"DBReadWriteAPI"` //gadget requests IDBReadWriteAPI
}



func (g *Foo) Run() {

	c := time.Tick(2 * time.Second)
	for now := range c {
		fmt.Printf("%v \n", now)

		if g.Settings != nil { //did the provider get set?
			if s,err := g.Settings.Keys("set1/") ; err == nil {
				fmt.Printf("Foo Setting Keys: %s \n", s)
			}
			if err := g.Settings.Put("set1/flagX", "bye"); err == nil {
			}
			if v,err := g.Settings.Get("set1/flagX"); err == nil {
				fmt.Printf("Foo Setting Value: %s \n", v)
			}
		}

		if g.DB != nil {
			if s,err := g.DB.Keys("/config/") ; err == nil {
				fmt.Printf("Foo DB Keys: %s \n", s)
			}
			if err := g.DB.Put("/config/DI!", "enabled"); err == nil {
			}
			if v,err := g.DB.Get("/config/DI!"); err == nil {

				fmt.Printf("Foo Your DB DI! settings are: %s \n", v)
			}

			fmt.Printf("Foo The following gadgets have stored settings (hopefully using ISettingsAPI)\n")
			if s,err := g.DB.Keys("/settings/") ; err == nil {
				if len(s) == 0 {
					fmt.Printf("Foo Hmmm, none yet!")
				} else {
					for _,v := range s {
						fmt.Printf("Foo Gadget: %s \n", v)
					}
				}
			}
		}
	}
}

func (g *Bar) Run() {

	c := time.Tick(5 * time.Second)
	for now := range c {
		fmt.Printf("%v \n", now)
		if g.Settings != nil {
			if s,err := g.Settings.Keys("") ; err == nil {
				fmt.Printf("Bar Setting Keys: %s \n", s)
			}
			if err := g.Settings.Put("Bar/Bar", "blacksheep"); err == nil {
			}
		}

		if g.DB != nil {
			if s,err := g.DB.Keys("sensor/") ; err == nil {
				fmt.Printf("Bar DB Keys: %s \n", s)
			}
		}
	}
}
