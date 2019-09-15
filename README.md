# easyconfig
A package for load config from env variables and yaml file easily.

# install
```bash
go get github.com/lianglee123/easyconfig
```

# example
```go
package main

import (
    "fmt"
    "os"
    "github.com/lianglee123/easyconfig"
)

type DBConfig struct{
	Host string   `config:"127.0.0.1"`
	Port int   `config:"5432"`
	UserName string `config:"lianglee"`
	Pwd string  `config:"qwer1234"`
	DBName string  `config:"config_demo"`
	Debug bool   `config:"-"` // not load this field
}

func main() {
	opt := &easyconfig.LoadOption{
		EnvPrefix: "DEMO",
		ConfigFilePath: "./test.yaml",
	}
	config := &DBConfig{}
	os.Setenv("DEMO_HOST", "111.111.111.000")
	err := easyconfig.LoadConfig(config, opt)
	if err != nil {
		fmt.Printf("err happen when load config: %v \n", err)
		return
	}
	fmt.Printf("config: %+v", config)
}
```
config file `test.yaml`:
```yaml
db:
  pwd: abcdefg
```

execute result:
```
config: &{Host:111.111.111.000 Port:5432 UserName:lianglee Pwd:abcdefg DBName:config_demo}
```
more detail [usage example](https://github.com/lianglee123/easyconfig_example/blob/master/main.go)  
# load priority
environment variables > yaml file > field tag(default value)

# Q & A
Q: How does easyconfig work?

A: easyconfig use viper under the hood. It has three steps: extra_default_values --> 
config viper instance --> set values to struct use viper.


Q: Why not use viper directly? 

A: Yes, you can. viper is a very good package, but i think its too complex in some scenes.