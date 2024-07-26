package cyi

import "github.com/gorilla/websocket"

func (cyi *Cyi) watch(id string, key string, conn *websocket.Conn) {
	_conn, _ := cyi.connList.Load(id)
	if cyi.channel[key] {
		_conn.(connKey).subscribe[key] = true
	} else {
		err := conn.WriteJSON([]any{
			key, resultCallError("cyi: key not string"),
		})
		if err != nil {
			conn.Close()
			return
		}
	}
}

func (cyi *Cyi) delWatch(id string, key string) {
	_conn, _ := cyi.connList.Load(id)
	delete(_conn.(connKey).subscribe, key)
}

func (cyi *Cyi) NewChannel(keys ...string) {
	for _, key := range keys {
		cyi.channel[key] = true
	}
}

func (cyi *Cyi) Send(id string, key string, data Result) {
	conn, ok := cyi.connList.Load(id)
	if ok && conn.(connKey).subscribe[key] {
		err := conn.(connKey).ws.WriteJSON([]any{
			key, data,
		})
		if err != nil {
			conn.(connKey).ws.Close()
			return
		}
	}
}
