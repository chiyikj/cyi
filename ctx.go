package cyi

type Ctx struct {
	Ip     string
	State  map[string]string
	Send   func(id string, key string, data Result)
	Plugin func(key string)any
}
