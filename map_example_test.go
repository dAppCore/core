package core_test

import . "dappco.re/go"

// ExampleMapKeys lists map keys through `MapKeys` for metadata maps. Map helpers cover
// selection, cloning, filtering, and merging with typed values.
func ExampleMapKeys() {
	ports := map[string]int{"api": 8080, "admin": 9090}
	keys := MapKeys(ports)
	SliceSort(keys)
	Println(keys)
	// Output: [admin api]
}

// ExampleMapValues lists map values through `MapValues` for metadata maps. Map helpers
// cover selection, cloning, filtering, and merging with typed values.
func ExampleMapValues() {
	ports := map[string]int{"api": 8080, "admin": 9090}
	values := MapValues(ports)
	SliceSort(values)
	Println(values)
	// Output: [8080 9090]
}

// ExampleMapClone clones a map through `MapClone` for metadata maps. Map helpers cover
// selection, cloning, filtering, and merging with typed values.
func ExampleMapClone() {
	original := map[string]int{"api": 8080}
	clone := MapClone(original)
	clone["api"] = 8181

	Println(original["api"])
	Println(clone["api"])
	// Output:
	// 8080
	// 8181
}

// ExampleMapFilter filters a map through `MapFilter` for metadata maps. Map helpers cover
// selection, cloning, filtering, and merging with typed values.
func ExampleMapFilter() {
	features := map[string]bool{"stable": true, "experimental": false}
	enabled := MapFilter(features, func(_ string, on bool) bool { return on })
	Println(MapHasKey(enabled, "stable"))
	Println(MapHasKey(enabled, "experimental"))
	// Output:
	// true
	// false
}

// ExampleMapMerge merges maps through `MapMerge` for metadata maps. Map helpers cover
// selection, cloning, filtering, and merging with typed values.
func ExampleMapMerge() {
	defaults := map[string]int{"port": 8080, "workers": 2}
	overrides := map[string]int{"workers": 4}
	merged := MapMerge(defaults, overrides)

	Println(merged["port"])
	Println(merged["workers"])
	// Output:
	// 8080
	// 4
}

// ExampleMapHasKey checks a map key through `MapHasKey` for metadata maps. Map helpers
// cover selection, cloning, filtering, and merging with typed values.
func ExampleMapHasKey() {
	Println(MapHasKey(map[string]int{"port": 8080}, "port"))
	// Output: true
}
