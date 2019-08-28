package config

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/magiconair/properties/assert"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"path/filepath"

	//"path/filepath"
	"reflect"
	//"strconv"
	"strings"
	"testing"
)

/*
1. 提取默认值（ExtraDefaultValues)
2. 实例化 viper， 并设置defaultValues (SetViperDefault)
3. 通过viper取回所有值 (LoadValue From Viper)

*/
type Person struct {
	Name   string `config:"name:ming_zi;default:joy"`
	Age    int64  `config:"default:24"`
	Logger struct {
		Path         string `config:"path"`
		Level        string `config:"level"`
		Env          string `config:"env"`
		ReportCaller bool   `config:"reportcaller"`
	} `config:"logger"`
}

func TestReflect(t *testing.T) {
	var person interface{} = &Person{}
	ptrVal := reflect.ValueOf(person)
	fmt.Println("type: ", ptrVal.Type())
	fmt.Println("kind: ", ptrVal.Kind())

	structVal := ptrVal.Elem()
	structType := structVal.Type()

	//defaultValues := map[string]string{}
	fmt.Println("structType PkgPath(): ", structType.PkgPath())
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)

		fmt.Printf("structField: %+v\n", structField)
		fmt.Printf("structField.Name: %+v\n", structField.Name)
		fmt.Printf("structFieldTag: %+v\n", structField.Tag.Get("config"))

		//configTag := structField.Tag.Get("config")
		////tagValues := TagToValues(configTag)
		//defaultValue, ok := tagValues["default"]
		//name, ok := tagValues["name"]
		//if !ok {
		//	name = strcase.ToSnake(structField.Name)
		//}
		//if ok {
		//	defaultValues[name] = defaultValue
		//}

	}

}

func LoadConfigFromViper(config interface{}, v *viper.Viper) (err error) {
	configPtrVal := reflect.ValueOf(config)
	configPtrType := configPtrVal.Type()

	configStructType := configPtrType.Elem()
	configStructVal := configPtrVal.Elem()

	for i := 0; i < configStructType.NumField(); i++ {
		field := configStructType.Field(i)
		fieldVal := configStructVal.Field(i)

		configTagStr := field.Tag.Get("config")
		values := ConfigTagToValues(configTagStr)

		var fieldKey string
		if _, ok := values["name"]; ok {
			fieldKey = values["name"]
		} else {
			fieldKey = strcase.ToSnake(field.Name)
		}
		viperKey := getFieldViperKey("", fieldKey)
		switch fieldVal.Kind() {
		case reflect.Int:
			fieldVal.SetInt(v.GetInt64(viperKey))
		case reflect.String:
			fieldVal.SetString(v.GetString(viperKey))
		}
	}
	return nil
}
func LoadConfig(config interface{}, opt *LoadOption) error {
	var envPrefix string
	var defaultConfigPath string
	if opt != nil {
		envPrefix = opt.EnvPrefix
		if opt.DefaultConfigPath != "" {
			defaultConfigPath = opt.DefaultConfigPath
		} else {
			defaultConfigPath = "."
		}
	}
	configPath := getConfigPath(envPrefix, defaultConfigPath)
	fmt.Println("configPath: ", configPath)
	configDir := filepath.Dir(configPath)
	configFile := strings.TrimSuffix(filepath.Base(configPath), ".yaml")

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(configFile)
	v.AddConfigPath(configDir)

	fmt.Println("configName: ", configFile, "configDir: ", configDir)
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	defaultValues := ExtraDefaultValues(config)
	SetViperDefault(v, defaultValues)
	fmt.Printf("default_values: %v", defaultValues)
	if err := v.ReadInConfig(); err != nil {
		fmt.Println("errOR: ", err)
		return err
	}

	if err := LoadConfigFromViper(config, v); err != nil {
		return err
	}
	return nil
}

func getFieldViperKey(prefix, fieldKey string) string {
	if prefix == "" {
		return fieldKey
	}
	return strings.Join([]string{prefix, fieldKey}, ".")
}

var finalKinds = map[reflect.Kind]bool{
	reflect.Bool:    true,
	reflect.Int:     true,
	reflect.Int8:    true,
	reflect.Int16:   true,
	reflect.Int32:   true,
	reflect.Int64:   true,
	reflect.Uint:    true,
	reflect.Uint8:   true,
	reflect.Uint16:  true,
	reflect.Uint32:  true,
	reflect.Uint64:  true,
	reflect.Float32: true,
	reflect.Float64: true,
	reflect.String:  true,

	//reflect.Array: true,
	//reflect.Slice: true,

}

var needRecursiveKinds = map[reflect.Kind]bool{
	reflect.Interface: true,
	reflect.Map:       true,
	reflect.Ptr:       true,
	reflect.Uintptr:   true,
	reflect.Struct:    true,
}

var unsupportKinds = map[reflect.Kind]bool{
	reflect.Complex64:     true,
	reflect.Complex128:    true,
	reflect.Chan:          true,
	reflect.Func:          true,
	reflect.UnsafePointer: true,
	reflect.Uintptr:       true,
}

// 如果没有明确设置默认，就设置为零值(指针除外)
// {'name': 'joy', 'logger.formater': 'json', 'logger.level': 'prod'}
func ExtraDefaultValues(config interface{}) map[string]interface{} {
	configPtrVal := reflect.ValueOf(config)
	configPtrType := configPtrVal.Type()

	configStructType := configPtrType.Elem()
	//configStructVal := configPtrVal.Elem()

	var result = make(map[string]interface{})

	for i := 0; i < configStructType.NumField(); i++ {
		field := configStructType.Field(i)
		//fieldVal := configPtrVal.Field(i)

		fieldType := field.Type
		fieldKind := fieldType.Kind()

		if _, ok := unsupportKinds[fieldKind]; ok {
			continue
		}

		//// ptr 迭代处理
		//if fieldKind == reflect.Ptr {
		//
		//}

		if _, ok := finalKinds[fieldKind]; ok {
			configTagStr := field.Tag.Get("config")
			values := ConfigTagToValues(configTagStr)
			var valueStr string
			if valueStr, ok = values["default"]; !ok {
				continue
			}

			var name string
			if _, ok = values["name"]; ok {
				name = values["name"]
			} else {
				name = strcase.ToSnake(field.Name)
			}
			value, err := StrConvertTo(valueStr, fieldKind)
			if err != nil {
				fmt.Printf("field %v default value(%v) has wrong type, \n Error: %v \n", field.Name, valueStr, err)
			}
			result[name] = value
		}
	}
	return result

}

//
func SetViperDefault(v *viper.Viper, values map[string]interface{}) {
	for key, val := range values {
		v.SetDefault(key, val)
	}
}

//
// tagStr format: key1:value1;key2:value2
func ConfigTagToValues(tagStr string) map[string]string {
	values := make(map[string]string)
	if tagStr == "" {
		return values
	}
	kvStrs := strings.Split(tagStr, ";")
	for _, kvStr := range kvStrs {
		if strings.Contains(kvStr, ":") {
			kvs := strings.Split(kvStr, ":")
			if len(kvs) >= 2 {
				k := strings.Trim(kvs[0], " ")
				v := kvs[1]
				values[k] = v
			}
		}
	}
	return values
}

func StrConvertTo(s string, kind reflect.Kind) (interface{}, error) {
	switch kind {
	case reflect.Float64:
		return cast.ToFloat64E(s)
	case reflect.Float32:
		return cast.ToFloat32E(s)
	case reflect.Int:
		return cast.ToIntE(s)
	case reflect.Int64:
		return cast.ToInt64E(s)
	case reflect.Int32:
		return cast.ToInt32E(s)
	case reflect.Int16:
		return cast.ToInt16E(s)
	case reflect.Int8:
		return cast.ToInt8E(s)
	case reflect.Uint:
		return cast.ToUintE(s)
	case reflect.Uint64:
		return cast.ToUint64E(s)
	case reflect.Uint32:
		return cast.ToUint32E(s)
	case reflect.Uint16:
		return cast.ToUint16E(s)
	case reflect.Uint8:
		return cast.ToUint8E(s)
	case reflect.String:
		return s, nil
	case reflect.Bool:
		if s == "true" {
			return true, nil
		} else {
			return false, nil
		}
	default:
		return nil, fmt.Errorf("unable to cast %#v to kind %T", s, kind)
	}
}

func TestExtraDefaultValues_1(t *testing.T) {
	type Person struct {
		Name  string `config:"default:joy"`
		Age   int    `config:"default:24"`
		Title string `config:"default:haha"`
	}
	p := &Person{}
	defaultValues := ExtraDefaultValues(p)
	fmt.Printf("%v", defaultValues)
	assert.Equal(t, defaultValues["name"], "joy")
	assert.Equal(t, defaultValues["age"], 24)
}

// test final type
func TestLoadConfig_1(t *testing.T) {
	type Person struct {
		Name string `config:"default:joy"`
		Age  int    `config:"default:24"`
	}
	p := &Person{}
	opt := &LoadOption{
		EnvPrefix:         "TEST",
		DefaultConfigPath: "./dev.yaml",
	}
	//os.Setenv("_TEST_AGE", "26")
	err := LoadConfig(p, opt)
	assert.Equal(t, err, nil)
	assert.Equal(t, p.Name, "joy")
	assert.Equal(t, p.Age, 28)
}

func TestLoadConfig_2(t *testing.T) {
	type Person struct {
		Name string `config:"default:joy"`
		Aag  *int
	}
	p := &Person{}
	opt := &LoadOption{
		EnvPrefix:         "TEST",
		DefaultConfigPath: "./dev.yaml",
	}
	err := LoadConfig(p, opt)
	assert.Equal(t, err, nil)
}
