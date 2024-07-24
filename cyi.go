package cyi

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

type request struct {
	Id            string
	MethodName    string
	ArgumentList  []any
	State         map[string]any
	errorMsg      string
	result        Result
	_ArgumentList []reflect.Value
}

var upgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type connKey struct {
	ws        *websocket.Conn
	subscribe map[string]bool
}

type Cyi struct {
	services     map[string]*method
	connList     *sync.Map
	channel      map[string]bool
	interceptors []func(method string, state map[string]string) *Result
	plugin       map[string]any
	openFunc     func(id string)
	closeFunc    func(id string)
}

type method struct {
	form          []form
	super         reflect.Type
	isInterceptor bool
}

type form struct {
	actualType  string
	entity      reflect.Type
	isNil       bool
	isBasicType bool
}

func New() *Cyi {
	return &Cyi{
		services: map[string]*method{},
		connList: &sync.Map{},
		channel:  make(map[string]bool),
		plugin:   make(map[string]any),
	}
}

func (cyi *Cyi) Start(addr uint16, ssl ...string) {
	http.HandleFunc("/", handleWebSocket(cyi))
	if len(ssl) == 2 {
		http.ListenAndServeTLS(":"+strconv.Itoa(int(addr)), ssl[0], ssl[1], nil)
	} else {
		http.ListenAndServe(":"+strconv.Itoa(int(addr)), nil)
	}
}

func (cyi *Cyi) SetPlugin(key string, value any) {
	cyi.plugin[key] = value
}
func (cyi *Cyi) OnOpen(callback func(id string)) {
	cyi.openFunc = callback
}

func (cyi *Cyi) OnClose(callback func(id string)) {
	cyi.closeFunc = callback
}

func (cyi *Cyi) Plugin(key string) any {
	return cyi.plugin[key]
}

func (cyi *Cyi) Bind(services ...any) {
	for _, service := range services {
		valueType := reflect.TypeOf(service)
		if valueType.Kind() == reflect.Struct {
			if valueType.NumField() != 1 || valueType.Field(0).Type != reflect.TypeOf(Ctx{}) || !valueType.Field(0).Anonymous {
				panic("The service parameter can be only one and is ctx and anonymous")
			}
			var isInterceptor = false
			for i := 0; i < valueType.NumMethod(); i++ {
				_method := valueType.Method(i)
				var methodName = valueType.Name() + "." + _method.Name
				// 获取方法的类型
				methodType := _method.Type
				if (_method.Name == "Interceptor" && methodType.NumOut() == 1 && methodType.Out(0).Kind() == reflect.Ptr && methodType.Out(0).Elem() == reflect.TypeOf(Result{}) && methodType.NumIn() == 2 &&
					methodType.In(1).Kind() == reflect.String) {
					isInterceptor = true
					continue
				}
				if methodType.NumOut() != 1 || methodType.Out(0) != reflect.TypeOf(Result{}) {
					panic("The service " + methodName + " needs only one return value and that is result")
				}
				// 遍历方法的参数类型
				var argumentList []form
				numOfParameters := methodType.NumIn()
				for j := 1; j < numOfParameters; j++ {
					parameterType := methodType.In(j)
					_form := form{}
					if !isType(parameterType) {
						panic("type " + parameterType.String() + " Not available")
					}
					if isComplexType(parameterType) {
						_form.isBasicType = false
					} else {
						_form.actualType = strings.Replace(parameterType.String(), "*", "", 1)
						_form.isBasicType = true
					}
					_form.isNil = parameterType.Kind() == reflect.Ptr
					if _form.isNil {
						_form.entity = parameterType.Elem()
					} else {
						_form.entity = parameterType
					}
					argumentList = append(argumentList, _form)
				}
				cyi.services[methodName] = &method{
					form:          argumentList,
					super:         valueType,
					isInterceptor: isInterceptor,
				}
			}
		} else {
			panic("The incoming service is not a struct")
		}
	}
}

func (cyi *Cyi) Interceptor(interceptors ...func(method string, state map[string]string) *Result) {
	for _, interceptor := range interceptors {
		cyi.interceptors = append(cyi.interceptors, interceptor)
	}
}

func handleWebSocket(cyi *Cyi) func(w http.ResponseWriter, r *http.Request) {
	closeFunc := func(conn *websocket.Conn, id string, status *bool) {
		if !*status {
			_statue := true
			status = &_statue
			_ = conn.Close()
			cyi.connList.Delete(id)
			if cyi.closeFunc != nil {
				cyi.closeFunc(id)
			}
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		id := r.URL.Query().Get("id")
		_statue := false
		status := &_statue
		defer func() {
			closeFunc(conn, id, status)
		}()
		if err != nil || r.URL.Query().Get("id") == "" {
			return
		}
		if cyi.openFunc != nil {
			cyi.openFunc(id)
		}
		var timer *time.Timer
		resetTimer := func() {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(10*time.Second, func() {
				closeFunc(conn, id, status)
			})
		}
		resetTimer()
		// 处理WebSocket连接
		ctx := Ctx{Ip: getIp(r), State: make(map[string]string), Send: cyi.Send, Plugin: cyi.Plugin, Id: id}
		cyi.connList.Store(id, connKey{ws: conn, subscribe: make(map[string]bool)})
		for {
			var request = &request{}
			var result Result
			err := conn.ReadJSON(request)
			if err != nil {
				var closeError *websocket.CloseError
				if errors.As(err, &closeError) {
					return
				}
				result = resultCallError("json: " + err.Error())
			} else if request.MethodName == "ping" {
				resetTimer()
				err = conn.WriteMessage(0, []byte("pong"))
				if err != nil {
					return
				}
				continue
			} else if request.Id == "" || request.MethodName == "" || request.ArgumentList == nil {
				result = resultCallError("json: cannot unmarshal number into Go value of type cyi.Request")
			} else {
				if request.MethodName == "watch" || request.MethodName == "delWatch" {
					handleWatch(cyi, id, conn, request, result)
					if request.errorMsg != "" {
						result = resultCallError(request.errorMsg)
					}
				} else {
					var method = cyi.services[request.MethodName]
					if method == nil {
						result = resultCallError("cyi: Method not found " + request.MethodName)
					} else if len(method.form) != len(request.ArgumentList) {
						result = resultCallError("cyi: ParameterMismatchError " + request.MethodName)
					} else {
						serialization(request, ctx.State, method.form)
						if request.errorMsg != "" {
							result = resultCallError(request.errorMsg)
						} else {
							cellMethod(request, method, ctx, cyi)
							if request.errorMsg != "" {
								result = resultCallError(request.errorMsg)
							} else {
								result = request.result
							}
						}
						ctx.State = make(map[string]string)
					}
				}
			}
			if (request.MethodName != "watch" && request.MethodName != "delWatch") || request.errorMsg != "" {
				result.Id = request.Id
				err = conn.WriteJSON(result)
				if err != nil {
					return
				}
			}
		}
	}
}

func handleWatch(cyi *Cyi, id string, conn *websocket.Conn, request *request, result Result) {
	key, ok := request.ArgumentList[0].(string)
	if !ok {
		request.errorMsg = "cyi: Invalid key format. Please provide a string key."
		return
	}
	if request.MethodName == "watch" {
		cyi.watch(id, key, request)
	} else {
		cyi.delWatch(id, key)
	}
}

// 执行方法
func cellMethod(request *request, _method *method, ctx Ctx, cyi *Cyi) {
	defer func() {
		if r := recover(); r != nil {
			request.errorMsg = fmt.Sprintf("%v", r)
		}
	}()
	newMethod, _ := _method.super.MethodByName(strings.Split(request.MethodName, ".")[1])
	newStruct := reflect.New(_method.super).Elem()
	field := newStruct.Field(0)
	field.Set(reflect.ValueOf(ctx))
	for _, interceptor := range cyi.interceptors {
		result := interceptor(request.MethodName, ctx.State)
		if result != nil {
			request.result = *result
			return
		}
	}
	if _method.isInterceptor {
		interceptor, _ := _method.super.MethodByName("Interceptor")
		result := interceptor.Func.Call([]reflect.Value{newStruct, reflect.ValueOf(request.MethodName)})[0].Interface().(*Result)
		if result != nil {
			request.result = *result
			return
		}
	}
	request._ArgumentList = append([]reflect.Value{newStruct}, request._ArgumentList...)
	request.result = newMethod.Func.Call(request._ArgumentList)[0].Interface().(Result)
}
