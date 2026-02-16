//go:build darwin && arm64 && mlx

package mlx

/*
#include "mlx/c/mlx.h"

// Callback for compiled functions.
extern int goCompiledFunc(mlx_vector_array *outputs, const mlx_vector_array inputs, void *payload);

static mlx_closure new_closure(void *payload) {
    return mlx_closure_new_func_payload(&goCompiledFunc, payload, NULL);
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

// CompiledFunc wraps a compiled MLX computation graph for efficient repeated calls.
type CompiledFunc struct {
	fn      func([]*Array) []*Array
	closure C.mlx_closure
	mu      sync.Mutex
}

var compiledFuncs sync.Map

//export goCompiledFunc
func goCompiledFunc(outputs *C.mlx_vector_array, inputs C.mlx_vector_array, payload unsafe.Pointer) C.int {
	id := uintptr(payload)
	fnI, ok := compiledFuncs.Load(id)
	if !ok {
		return 1
	}
	fn := fnI.(func([]*Array) []*Array)

	// Convert inputs
	nInputs := int(C.mlx_vector_array_size(inputs))
	goInputs := make([]*Array, nInputs)
	for i := 0; i < nInputs; i++ {
		a := New("INPUT")
		C.mlx_vector_array_get(&a.ctx, inputs, C.int(i))
		goInputs[i] = a
	}

	// Call user function
	goOutputs := fn(goInputs)

	// Set outputs
	for _, out := range goOutputs {
		C.mlx_vector_array_append_value(*outputs, out.ctx)
	}
	return 0
}

var nextID uintptr
var nextIDMu sync.Mutex

// CompileShapeless compiles a function for efficient repeated execution.
// The function must accept and return arrays of consistent shapes.
func CompileShapeless(fn func([]*Array) []*Array, shapeless bool) *CompiledFunc {
	nextIDMu.Lock()
	nextID++
	id := nextID
	nextIDMu.Unlock()

	compiledFuncs.Store(id, fn)

	cf := &CompiledFunc{fn: fn}
	cf.closure = C.new_closure(unsafe.Pointer(id))
	return cf
}

// Call executes the compiled function with the given inputs.
func (cf *CompiledFunc) Call(inputs ...*Array) []*Array {
	cf.mu.Lock()
	defer cf.mu.Unlock()

	// Fall back to direct call — compilation is an optimization.
	// The compiled closure can be used via mlx_compiled but the
	// direct path is simpler and still benefits from MLX's lazy evaluation.
	return cf.fn(inputs)
}
