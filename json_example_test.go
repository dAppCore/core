package core_test

import . "dappco.re/go"

// ExampleJSONMarshal_config initialises configuration values through `JSONMarshal` for
// configuration serialisation. Serialisation and parsing return core Results for
// configuration payloads.
func ExampleJSONMarshal_config() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	r := JSONMarshal(appConfig{Host: "localhost", Port: 8080})
	Println(string(r.Value.([]byte)))
	// Output: {"host":"localhost","port":8080}
}

// ExampleJSONMarshalString serialises a value to JSON text through `JSONMarshalString` for
// configuration serialisation. Serialisation and parsing return core Results for
// configuration payloads.
func ExampleJSONMarshalString() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	Println(JSONMarshalString(appConfig{Host: "localhost", Port: 8080}))
	// Output: {"host":"localhost","port":8080}
}

// ExampleJSONUnmarshal parses JSON bytes through `JSONUnmarshal` for configuration
// serialisation. Serialisation and parsing return core Results for configuration payloads.
func ExampleJSONUnmarshal() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	var cfg appConfig
	JSONUnmarshal([]byte(`{"host":"localhost","port":8080}`), &cfg)
	Println(cfg.Host, cfg.Port)
	// Output: localhost 8080
}

// ExampleJSONUnmarshalString_config initialises configuration values through
// `JSONUnmarshalString` for configuration serialisation. Serialisation and parsing return
// core Results for configuration payloads.
func ExampleJSONUnmarshalString_config() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	var cfg appConfig
	JSONUnmarshalString(`{"host":"localhost","port":8080}`, &cfg)
	Println(cfg.Host, cfg.Port)
	// Output: localhost 8080
}
