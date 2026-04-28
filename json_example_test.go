package core_test

import . "dappco.re/go"

func ExampleJSONMarshal_config() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	r := JSONMarshal(appConfig{Host: "localhost", Port: 8080})
	Println(string(r.Value.([]byte)))
	// Output: {"host":"localhost","port":8080}
}

func ExampleJSONMarshalString() {
	type appConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	Println(JSONMarshalString(appConfig{Host: "localhost", Port: 8080}))
	// Output: {"host":"localhost","port":8080}
}

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
