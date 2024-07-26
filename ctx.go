package cyi

type Ctx struct {
	Ip     string
	Id     string
	State  map[string]string
	Send   func(id string, key string, data Result) bool
	Plugin func(key string) any
}
