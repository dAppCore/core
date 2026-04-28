package core_test

import . "dappco.re/go"

func ExampleMapKeys() {
	ports := map[string]int{"api": 8080, "admin": 9090}
	keys := MapKeys(ports)
	SliceSort(keys)
	Println(keys)
	// Output: [admin api]
}

func ExampleMapValues() {
	ports := map[string]int{"api": 8080, "admin": 9090}
	values := MapValues(ports)
	SliceSort(values)
	Println(values)
	// Output: [8080 9090]
}

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

func ExampleMapFilter() {
	features := map[string]bool{"stable": true, "experimental": false}
	enabled := MapFilter(features, func(_ string, on bool) bool { return on })
	Println(MapHasKey(enabled, "stable"))
	Println(MapHasKey(enabled, "experimental"))
	// Output:
	// true
	// false
}

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

func ExampleMapHasKey() {
	Println(MapHasKey(map[string]int{"port": 8080}, "port"))
	// Output: true
}
