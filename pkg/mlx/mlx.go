//go:build darwin && arm64 && mlx

// Package mlx provides Go bindings for Apple's MLX framework via mlx-c.
//
// Build mlx-c before use:
//
//	cd pkg/mlx && go generate ./...
//
// Build with MLX enabled:
//
//	go build -tags mlx -o core .
package mlx

//go:generate cmake -S . -B build -DCMAKE_INSTALL_PREFIX=dist -DCMAKE_BUILD_TYPE=Release
//go:generate cmake --build build --parallel
//go:generate cmake --install build

/*
#cgo CXXFLAGS: -std=c++17
#cgo CPPFLAGS: -I${SRCDIR}/dist/include
#cgo LDFLAGS: -L${SRCDIR}/dist/lib -lmlxc -lmlx -lstdc++
#cgo darwin LDFLAGS: -framework Foundation -framework Metal -framework Accelerate
#cgo darwin LDFLAGS: -Wl,-rpath,${SRCDIR}/dist/lib

#include <stdlib.h>
#include "mlx/c/mlx.h"

static const char *last_mlx_error = NULL;

static void mlx_go_error_handler(const char *msg, void *data) {
    last_mlx_error = msg;
}

static void set_error_handler() {
    mlx_set_error_handler(&mlx_go_error_handler, NULL, NULL);
}

static const char* get_last_error() {
    return last_mlx_error;
}
*/
import "C"

import (
	"log/slog"
	"sync"
)

var initOnce sync.Once

// Init sets up the MLX error handler. Called automatically on first use.
func Init() {
	initOnce.Do(func() {
		C.set_error_handler()
		slog.Debug("mlx: initialized with Metal backend")
	})
}

// checkError logs the last MLX error if any occurred.
func checkError() {
	if msg := C.get_last_error(); msg != nil {
		slog.Error("mlx", "error", C.GoString(msg))
	}
}

// Materialize synchronously evaluates arrays, computing their values on the GPU.
// This is the MLX equivalent of forcing lazy computation to complete.
func Materialize(outputs ...*Array) {
	doMaterialize(outputs, false)
}

// MaterializeAsync queues arrays for asynchronous GPU evaluation.
func MaterializeAsync(outputs ...*Array) {
	doMaterialize(outputs, true)
}

func doMaterialize(outputs []*Array, async bool) {
	Init()
	vector := C.mlx_vector_array_new()
	defer C.mlx_vector_array_free(vector)

	for _, output := range outputs {
		if output != nil && output.Valid() {
			C.mlx_vector_array_append_value(vector, output.ctx)
		}
	}

	if async {
		C.mlx_async_eval(vector)
	} else {
		C.mlx_eval(vector)
	}
}

// Collect gathers all valid arrays from a variadic list for batch Materialize.
func Collect(arrays ...*Array) []*Array {
	var out []*Array
	for _, a := range arrays {
		if a != nil && a.Valid() {
			out = append(out, a)
		}
	}
	return out
}

// MetalAvailable reports whether Metal GPU is available.
func MetalAvailable() bool {
	Init()
	var available C.bool
	C.mlx_metal_is_available(&available)
	return bool(available)
}
