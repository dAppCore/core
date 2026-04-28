package core_test

import . "dappco.re/go"

// ExampleRegex compiles a pattern through `Regex` for route parsing. Compiled patterns
// expose common matching and replacement operations through core wrappers.
func ExampleRegex() {
	r := Regex(`\d+`)
	rx := r.Value.(*Regexp)

	Println(rx.MatchString("build 42"))
	Println(rx.FindString("build 42"))
	// Output:
	// true
	// 42
}

// ExampleRegexp_FindAllString finds all string matches through `Regexp.FindAllString` for
// route parsing. Compiled patterns expose common matching and replacement operations
// through core wrappers.
func ExampleRegexp_FindAllString() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.FindAllString("a1 b22 c333", -1))
	// Output: [1 22 333]
}

// ExampleRegexp_FindStringSubmatch captures submatches through `Regexp.FindStringSubmatch`
// for route parsing. Compiled patterns expose common matching and replacement operations
// through core wrappers.
func ExampleRegexp_FindStringSubmatch() {
	rx := Regex(`(\w+)=(\d+)`).Value.(*Regexp)
	Println(rx.FindStringSubmatch("port=8080"))
	// Output: [port=8080 port 8080]
}

// ExampleRegexp_ReplaceAllString replaces matches through `Regexp.ReplaceAllString` for
// route parsing. Compiled patterns expose common matching and replacement operations
// through core wrappers.
func ExampleRegexp_ReplaceAllString() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.ReplaceAllString("api:8080 admin:9090", "PORT"))
	// Output: api:PORT admin:PORT
}

// ExampleRegexp_Split splits text through `Regexp.Split` for route parsing. Compiled
// patterns expose common matching and replacement operations through core wrappers.
func ExampleRegexp_Split() {
	rx := Regex(`,+`).Value.(*Regexp)
	Println(rx.Split("alpha,bravo,,charlie", -1))
	// Output: [alpha bravo charlie]
}

// ExampleRegexp_String renders `Regexp.String` as a stable string for route parsing.
// Compiled patterns expose common matching and replacement operations through core
// wrappers.
func ExampleRegexp_String() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.String())
	// Output: \d+
}
