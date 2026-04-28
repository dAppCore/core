package core_test

import . "dappco.re/go"

func ExampleSHA256() {
	sum := SHA256([]byte("hello"))
	Println(HexEncode(sum[:]))
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}

func ExampleSHA256Hex() {
	Println(SHA256Hex([]byte("hello")))
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}

func ExampleSHA256String() {
	sum := SHA256String("hello")
	Println(HexEncode(sum[:]))
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}

func ExampleSHA256HexString() {
	Println(SHA256HexString("hello"))
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}

func ExampleHMAC() {
	r := HMAC("sha256", []byte("secret"), []byte("payload"))
	if r.OK {
		digest := r.Value.([]byte)
		Println(len(digest))
		Println(HexEncode(digest)[:8])
	}
	// Output:
	// 32
	// b82fcb79
}

func ExampleHKDF() {
	r := HKDF("sha256", []byte("secret"), []byte("salt"), []byte("session"), 32)
	if r.OK {
		key := r.Value.([]byte)
		Println(len(key))
		Println(HexEncode(key)[:8])
	}
	// Output:
	// 32
	// 2baf2709
}
