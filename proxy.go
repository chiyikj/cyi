package cyi

import "github.com/gorilla/websocket"

func (cyi *Cyi) watch(id string, keys []string, conn *websocket.Conn) {
	_conn, _ := cyi.connList.Load(id)
	for _, key := range keys {
		if cyi.channel[key] {
			_conn.(connKey).subscribe[key] = true
		} else {
			for _, key := range keys {
				if cyi.channel[key] {
					delete(_conn.(connKey).subscribe, key)
				}
			}
			err := conn.WriteJSON([]any{
				id, resultCallError("cyi: The server does not have a subscription for this key"),
			})
			if err != nil {
				conn.Close()
				return
			}
			return
		}
	}
}

func (cyi *Cyi) delWatch(id string, keys []string) {
	_conn, _ := cyi.connList.Load(id)
	for _, key := range keys {
		delete(_conn.(connKey).subscribe, key)
	}
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
