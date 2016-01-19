package common

import (
	"reflect"
)

type Syscall struct {
	Name     string
	Kernel   *KernelBase
	Instance reflect.Value
	Method   reflect.Method
	In       []reflect.Type
	Out      []reflect.Type
	ObufArr  bool
}

// Call a syscall from the dispatch table. Will panic() if anything goes terribly wrong.
func (sys Syscall) Call(args []uint64) uint64 {
	extraArgs := 1
	if sys.ObufArr {
		extraArgs += 1
	}
	in := make([]reflect.Value, len(sys.In)+extraArgs)
	in[0] = sys.Instance
	// special case "all args" buf list
	if sys.ObufArr {
		arr := make([]Obuf, len(sys.In)-1)
		for i := range arr {
			arr[i] = Obuf{NewBuf(sys.Kernel.U, args[i])}
		}
		in[1] = reflect.ValueOf(arr)
	}
	// convert syscall arguments
	converted, err := sys.Kernel.Argjoy.Convert(sys.In, false, args)
	if err != nil {
		panic(err)
	}
	copy(in[extraArgs:], converted)
	// call handler function
	out := sys.Method.Func.Call(in)
	// return output if first return of function is representable as an int type
	Uint64Type := reflect.TypeOf(uint64(0))
	if len(out) > 0 && out[0].Type().ConvertibleTo(Uint64Type) {
		return out[0].Convert(Uint64Type).Uint()
	}
	return 0
}
