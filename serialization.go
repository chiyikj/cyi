package cyi

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

func serialization(request *request, state map[string]string, from []form) {
	defer func() {
		if r := recover(); r != nil {
			request.errorMsg = fmt.Sprintf("%v", r)
		}
	}()
	for index, value := range request.State {
		state[index] = value.(string)
	}
	for index, value := range request.ArgumentList {
		if value == nil {
			if from[index].isNil {
				request._ArgumentList = append(request._ArgumentList, reflect.Zero(reflect.PointerTo(from[index].entity)))
			} else {
				panic("argument " + strconv.Itoa(index) + from[index].actualType + " no nil")
			}
		} else {
			if from[index].isBasicType {
				_value := reflect.New(from[index].entity).Elem()
				if from[index].actualType != "int" {
					_value.Set(reflect.ValueOf(value))
				} else {
					__value := value.(float64)
					if __value > math.MaxInt || __value < math.MinInt {
						panic("argument " + strconv.Itoa(index) + from[index].actualType + " out of range")
					}
					_value.Set(reflect.ValueOf(int(__value)))
				}
				if from[index].isNil {
					request._ArgumentList = append(request._ArgumentList, _value.Addr())
				} else {
					request._ArgumentList = append(request._ArgumentList, _value)
				}
			} else {
				_json, _ := json.Marshal(value)
				newStruct := reflect.New(from[index].entity)
				_ = json.Unmarshal(_json, newStruct.Interface())
				if from[index].isNil {
					request._ArgumentList = append(request._ArgumentList, newStruct)
				} else {
					request._ArgumentList = append(request._ArgumentList, newStruct.Elem())
				}
			}
		}
	}
}
