package core_test

import . "dappco.re/go"

func ExampleRegex() {
	r := Regex(`\d+`)
	rx := r.Value.(*Regexp)

	Println(rx.MatchString("build 42"))
	Println(rx.FindString("build 42"))
	// Output:
	// true
	// 42
}

func ExampleRegexp_FindAllString() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.FindAllString("a1 b22 c333", -1))
	// Output: [1 22 333]
}

func ExampleRegexp_FindStringSubmatch() {
	rx := Regex(`(\w+)=(\d+)`).Value.(*Regexp)
	Println(rx.FindStringSubmatch("port=8080"))
	// Output: [port=8080 port 8080]
}

func ExampleRegexp_ReplaceAllString() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.ReplaceAllString("api:8080 admin:9090", "PORT"))
	// Output: api:PORT admin:PORT
}

func ExampleRegexp_Split() {
	rx := Regex(`,+`).Value.(*Regexp)
	Println(rx.Split("alpha,bravo,,charlie", -1))
	// Output: [alpha bravo charlie]
}

func ExampleRegexp_String() {
	rx := Regex(`\d+`).Value.(*Regexp)
	Println(rx.String())
	// Output: \d+
}
