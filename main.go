package main

import (
	. "github.com/pontaoski/goobject/libgoobject"
)

func init() {
	RegisterClass("Test", "Example", GOObjectClass{
		AllowInherit: true,
		Signals:      Signals{"oof"},
		Properties: Properties{
			"color": Int64,
		},
		Functions: Functions{
			"yeet": func() {
				println("yeet")
			},
		},
	})
	RegisterInheritedClass("Test", "Example", "Test", "Ouch", GOObjectClass{
		AllowInherit: false,
		Signals:      Signals{"yeet"},
		Properties: Properties{
			"color_alt": Int64,
		},
	})
}

func main() {
	emitter := ConstructGOObject("Test", "Example")
	channel := make(Signal)
	emitter.ConnectSignal("notify", channel)
	go func() {
		<-channel
		println("properties updated")
	}()
	emitter.Set("color", 5)
}
