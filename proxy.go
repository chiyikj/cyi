package cyi

import "github.com/gorilla/websocket"

func (cyi *Cyi) watch(id string, keys []string, conn *websocket.Conn, requestId string, joinId string) {
	_conn, ok := cyi.connList.Load(id)
	if ok {
		connMap, _ := _conn.(map[string]*connKey)
		if connMap[joinId] != nil {
			for _, key := range keys {
				if cyi.channel[key] {
					connMap[joinId].subscribe[key] = true
				} else {
					for _, key := range keys {
						if cyi.channel[key] {
							delete(connMap[joinId].subscribe, key)
						}
					}
					err := conn.WriteJSON([]any{
						requestId, resultCallError("cyi: The server does not have a subscription for this key"),
					})
					if err != nil {
						conn.Close()
						return
					}
					return
				}
			}
		}
	}
}

func (cyi *Cyi) delWatch(id string, keys []string, joinId string) {
	_conn, ok := cyi.connList.Load(id)
	if ok {
		connMap, _ := _conn.(map[string]*connKey)
		if connMap[joinId] != nil {
			for _, key := range keys {
				delete(connMap[joinId].subscribe, key)
			}
		}
	}
}

func (cyi *Cyi) NewChannel(keys ...string) {
	for _, key := range keys {
		cyi.channel[key] = true
	}
}

func (cyi *Cyi) Send(id string, key string, data Result) bool {
	conn, ok := cyi.connList.Load(id)
	if ok {
		connMap, _ := conn.(map[string]*connKey)
		for _, _conn := range connMap {
			if _conn.subscribe[key] {
				err := _conn.ws.WriteJSON([]any{
					key, data,
				})
				if err != nil {
					_conn.ws.Close()
					return false
				} else {
					return true
				}
			}
		}
	}
	return false
}
