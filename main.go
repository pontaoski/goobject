package main

import (
	"reflect"

	. "github.com/pontaoski/goobject/libgoobject"
)

func init() {
	RegisterClass("Test", "Example", GOObjectClass{
		AllowInherit: true,
		Signals:      Signals{"oof"},
		Properties: Properties{
			"color": reflect.TypeOf(0),
		},
		Functions: Functions{
			"yeet": func() {
				println("yeet")
			},
		},
	})
	RegisterInheritedClass("Test", "Example", "Test", "Ouch", GOObjectClass{})
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
