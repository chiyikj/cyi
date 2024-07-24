package cyi

func (cyi *Cyi) watch(id string, key string, request *request) {
	_conn, _ := cyi.connList.Load(id)
	if cyi.channel[key] {
		_conn.(connKey).subscribe[key] = true
	} else {
		request.errorMsg = "cyi: The server does not have a subscription for this key"
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
