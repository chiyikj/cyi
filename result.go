package cyi

type Result struct {
	Id   string
	Code  *uint16
	Msg   string
	Data interface{}
}

func ResultSuccess(data interface{}) Result {
	return Result{Data: data}
}

func ResultError(code uint16, msg string) Result {
	return Result{Code: &code,Msg: msg}
}

func resultCallError(msg string) Result {
	return Result{Msg: msg}
}
