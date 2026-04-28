package core_test

import . "dappco.re/go"

func Example_sha3_256() {
	sum := SHA3_256([]byte("hello"))
	Println(HexEncode(sum[:])[:16])
	// Output: 3338be694f50c5f3
}

func Example_sha3_256Hex() {
	Println(SHA3_256Hex([]byte("hello"))[:16])
	// Output: 3338be694f50c5f3
}

func ExampleKeccak256() {
	sum := Keccak256([]byte("hello"))
	Println(HexEncode(sum[:])[:16])
	// Output: 1c8aff950685c2ed
}

func ExampleKeccak256Hex() {
	Println(Keccak256Hex([]byte("hello"))[:16])
	// Output: 1c8aff950685c2ed
}

func ExampleSHA3Shake128() {
	sum := SHA3Shake128([]byte("hello"), 8)
	Println(len(sum))
	Println(HexEncode(sum))
	// Output:
	// 8
	// 8eb4b6a932f28033
}

func ExampleSHA3Shake256() {
	sum := SHA3Shake256([]byte("hello"), 8)
	Println(len(sum))
	Println(HexEncode(sum))
	// Output:
	// 8
	// 1234075ae4a1e773
}
