package goobject

import (
	"encoding/json"
	"log"
	"reflect"
)

type GOObjectClass struct {
	Constructor  func(g *GOObject)
	Functions    map[string]interface{}
	Properties   map[string]reflect.Type
	Signals      []string
	AllowInherit bool
	inherit      *GOObjectClass
}

type GOObject struct {
	properties map[string]interface{}
	signals    map[string][]Signal
	class      GOObjectClass
}

var registeredClasses = make(map[string]map[string]GOObjectClass)

type Signal chan struct{}
type Signals []string
type Properties map[string]reflect.Type
type Functions map[string]interface{}

func output(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "\t")
	println(string(data))
}

func checkRegistered(registerNamespace, registerName string) {
	for namespaceName, namespace := range registeredClasses {
		for className := range namespace {
			if namespaceName == registerNamespace && className == registerName {
				log.Fatalf("Class %s:%s is already registered", namespaceName, registerName)
			}
		}
	}
}

func locateClass(namespace, name string) GOObjectClass {
	for namespaceName, namespace := range registeredClasses {
		for className, class := range namespace {
			if namespaceName == namespaceName && className == name {
				return class
			}
		}
	}
	log.Fatalf("Class %s:%s not found", namespace, name)
	return GOObjectClass{}
}

func classHasSignal(class GOObjectClass, signal string) bool {
	parent := &class
	for parent != nil {
		for _, classSignal := range parent.Signals {
			if classSignal == signal {
				return true
			}
		}
		parent = parent.inherit
	}
	return false
}

func classHasProperty(class GOObjectClass, property string) bool {
	parent := &class
	for parent != nil {
		for classProperty := range parent.Properties {
			if classProperty == property {
				return true
			}
		}
		parent = parent.inherit
	}
	return false
}

func classPropertyType(class GOObjectClass, property string) reflect.Type {
	parent := &class
	for parent != nil {
		for classProperty, propertyType := range parent.Properties {
			if classProperty == property {
				return propertyType
			}
		}
		parent = parent.inherit
	}
	return nil
}

func init() {
	if registeredClasses["libgoobject"] == nil {
		registeredClasses["libgoobject"] = make(map[string]GOObjectClass)
	}
	registeredClasses["libgoobject"]["goobject"] = GOObjectClass{AllowInherit: true, Signals: Signals{"notify"}}
}

func RegisterInheritedClass(parentNamespace, parentName, namespace, name string, class GOObjectClass) {
	checkRegistered(namespace, name)
	if registeredClasses[namespace] == nil {
		registeredClasses[namespace] = make(map[string]GOObjectClass)
	}
	parent := locateClass(parentNamespace, parentName)
	if !parent.AllowInherit {
		log.Fatalf("Cannot inherit from class %s:%s", parentNamespace, parentName)
	}
	class.inherit = &parent
	registeredClasses[namespace][name] = class
}

func RegisterClass(namespace, name string, class GOObjectClass) {
	checkRegistered(namespace, name)
	if registeredClasses[namespace] == nil {
		registeredClasses[namespace] = make(map[string]GOObjectClass)
	}
	root := registeredClasses["libgoobject"]["goobject"]
	class.inherit = &root
	registeredClasses[namespace][name] = class
}

func ConstructGOObject(namespaceType, classType string) *GOObject {
	for namespaceName, namespace := range registeredClasses {
		for className, class := range namespace {
			if namespaceName == namespaceType && className == classType {
				goobject := &GOObject{class: class}
				v := &class
				for v != nil {
					if v.Constructor != nil {
						v.Constructor(goobject)
						break
					}
					v = v.inherit
				}
				return goobject
			}
		}
	}
	return nil
}

func (g *GOObjectClass) Super() *GOObjectClass {
	return g.inherit
}

func (g *GOObject) Call(function string, v ...interface{}) interface{} {
	var located interface{}
	class := &g.class
	for class != nil {
		if class.Functions != nil {
			if val, ok := class.Functions[function]; ok {
				located = val
			}
		}
		class = class.inherit
	}
	val := reflect.ValueOf(located)
	if val.Type().Kind() != reflect.Func {
		log.Fatalf("Function %s is not a function", function)
	}
	var values []reflect.Value
	for _, val := range v {
		values = append(values, reflect.ValueOf(val))
	}
	result := val.Call(values)
	var interfaces []interface{}
	for _, val := range result {
		interfaces = append(interfaces, val.Interface())
	}
	return interfaces
}

func (g *GOObject) EmitSignal(signal string) {
	if !classHasSignal(g.class, signal) {
		log.Fatalf("Signal %s does not exist", signal)
	}
	if val, ok := g.signals[signal]; ok {
		for _, listener := range val {
			listener <- struct{}{}
		}
	}
}

func (g *GOObject) ConnectSignal(name string, signal Signal) {
	if !classHasSignal(g.class, name) {
		log.Fatalf("Signal %s does not exist", signal)
	}
	if g.signals == nil {
		g.signals = make(map[string][]Signal)
	}
	val, ok := g.signals[name]
	if !ok {
		g.signals[name] = []Signal{signal}
	} else {
		g.signals[name] = append(val, signal)
	}
}

func (g *GOObject) Get(property string) interface{} {
	if !classHasProperty(g.class, property) {
		log.Fatalf("Property %s does not exist", property)
	}
	return g.properties[property]
}

func (g *GOObject) Set(property string, value interface{}) {
	propertyType := classPropertyType(g.class, property)
	if reflect.TypeOf(value) != propertyType {
		log.Fatalf("A value of type %s is not compatible with property %s of type %s", reflect.TypeOf(value).String(), property, propertyType.String())
	}
	if g.properties == nil {
		g.properties = make(map[string]interface{})
	}
	g.EmitSignal("notify")
	g.properties[property] = value
}
