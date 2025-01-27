//go:build js

package syscall

import (
	"syscall/js"
)

// fsCall emulates a file system-related syscall via a corresponding NodeJS fs
// API.
//
// This version is similar to the upstream, but it gracefully handles missing fs
// methods (allowing for smaller prelude) and removes a workaround for an
// obsolete NodeJS version.
func fsCall(name string, args ...interface{}) (js.Value, error) {
	type callResult struct {
		val js.Value
		err error
	}

	c := make(chan callResult, 1)
	f := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var res callResult

		if jsErr := args[0]; !jsErr.IsNull() {
			res.err = mapJSError(jsErr)
		}

		res.val = js.Undefined()
		if len(args) >= 2 {
			res.val = args[1]
		}

		c <- res
		return nil
	})
	defer f.Release()
	if jsFS.Get(name).IsUndefined() {
		return js.Undefined(), ENOSYS
	}
	jsFS.Call(name, append(args, f)...)
	res := <-c
	return res.val, res.err
}
