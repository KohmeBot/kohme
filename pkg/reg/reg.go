package reg

import (
	"fmt"
	"github.com/kohmebot/kohme"
	"github.com/kohmebot/plugin/v2"
	"reflect"
)

func RegisterAny(v any) {
	var p plugin.Plugin
	switch t := v.(type) {
	case func() plugin.Plugin:
		p = t()
	case plugin.Plugin:
		p = t
	default:
		rv := reflect.ValueOf(v)

		if rv.Kind() != reflect.Func {
			break
		}
		res := rv.Call([]reflect.Value{})
		for _, rv = range res {
			if !rv.CanInterface() || rv.IsNil() {
				continue
			}
			var ok bool
			p, ok = rv.Interface().(plugin.Plugin)
			if ok {
				break
			}
		}
	}

	if p == nil {
		panic(fmt.Errorf("can not register %T", v))
	}

	kohme.Register(p)
}
