//go:build !(darwin && arm64 && mlx)

// Package mlx provides Go bindings for Apple's MLX framework via mlx-c.
// This stub file is used on non-darwin/non-arm64 platforms or when the
// mlx build tag is not set. All operations report MLX as unavailable.
package mlx

// MetalAvailable reports whether Metal GPU is available.
// Always returns false on non-Apple Silicon platforms.
func MetalAvailable() bool { return false }
