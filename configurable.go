package configurable

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
	"gopkg.in/yaml.v3"
)

// IConfigurable defines the interface for configuration management.
type IConfigurable interface {
	// Existing methods
	Int(name string) *int
	NewInt(name string, value int, usage string) *int

	Int64(name string) *int64
	NewInt64(name string, value int64, usage string) *int64

	Float64(name string) *float64
	NewFloat64(name string, value float64, usage string) *float64

	String(name string) *string
	NewString(name, value, usage string) *string

	Bool(name string) *bool
	NewBool(name string, value bool, usage string) *bool

	Duration(name string) *time.Duration
	NewDuration(name string, value time.Duration, usage string) *time.Duration

	List(name string) *[]string
	NewList(name string, value []string, usage string) *[]string

	Map(name string) *map[string]string
	NewMap(name string, value map[string]string, usage string) *map[string]string

	LoadFile(filename string) error
	Parse(filename string) error

	Usage() string
}

type Configurable struct {
	flags map[string]interface{}
}

func New() IConfigurable {
	return &Configurable{flags: make(map[string]interface{})}
}

func (c *Configurable) NewInt(name string, value int, usage string) *int {
	ptr := flag.Int(name, value, usage)
	c.flags[name] = ptr
	return ptr
}

func (c *Configurable) Int(name string) *int {
	c.checkAndSetFromEnv(name)
	if ptr, ok := c.flags[name].(*int); ok {
		return ptr
	}
	return nil
}

func (c *Configurable) Int64(name string) *int64 {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*int64)
	return val
}

func (c *Configurable) NewInt64(name string, value int64, usage string) *int64 {
	var i = flag.Int64(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) Float64(name string) *float64 {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*float64)
	return val
}

func (c *Configurable) NewFloat64(name string, value float64, usage string) *float64 {
	var i = flag.Float64(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) Duration(name string) *time.Duration {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*time.Duration)
	return val
}

func (c *Configurable) NewDuration(name string, value time.Duration, usage string) *time.Duration {
	var i = flag.Duration(name, value, usage)
	c.flags[name] = i
	return i
}

func (c *Configurable) String(name string) *string {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*string)
	return val
}

func (c *Configurable) NewString(name string, value string, usage string) *string {
	var s = flag.String(name, value, usage)
	c.flags[name] = s
	return s
}

func (c *Configurable) Bool(name string) *bool {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*bool)
	return val
}

func (c *Configurable) NewBool(name string, value bool, usage string) *bool {
	var b = flag.Bool(name, value, usage)
	c.flags[name] = b
	return b
}

type ListFlag struct {
	values *[]string
}

func (l *ListFlag) String() string {
	if l.values == nil {
		return ""
	}
	return strings.Join(*l.values, ",")
}

func (l *ListFlag) Set(value string) error {
	if l.values == nil {
		l.values = &[]string{}
	}
	items := strings.Split(value, ",")
	*l.values = append(*l.values, items...)
	return nil
}

func (c *Configurable) NewList(name string, value []string, usage string) *[]string {
	l := &ListFlag{values: &value}
	flag.Var(l, name, usage)
	c.flags[name] = l
	return l.values
}

func (c *Configurable) List(name string) *[]string {
	c.checkAndSetFromEnv(name)
	if ptr, ok := c.flags[name].(*ListFlag); ok {
		return ptr.values
	}
	return nil
}

type MapFlag struct {
	values *map[string]string
}

func (m *MapFlag) String() string {
	if m.values == nil {
		return ""
	}
	var entries []string
	for k, v := range *m.values {
		entries = append(entries, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(entries, ",")
}

func (m *MapFlag) Set(value string) error {
	if m.values == nil {
		m.values = &map[string]string{}
	}
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid map item: %s", pair)
		}
		(*m.values)[kv[0]] = kv[1]
	}
	return nil
}

func (c *Configurable) NewMap(name string, value map[string]string, usage string) *map[string]string {
	m := &MapFlag{values: &value}
	flag.Var(m, name, usage)
	c.flags[name] = m
	return m.values
}

func (c *Configurable) Map(name string) *map[string]string {
	c.checkAndSetFromEnv(name)
	if ptr, ok := c.flags[name].(*MapFlag); ok {
		return ptr.values
	}
	return nil
}

func (c *Configurable) Parse(filename string) error {
	flag.Parse()
	if filename != "" {
		return c.LoadFile(filename)
	}
	return nil
}

func (c *Configurable) LoadFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return c.loadJSON(data)
	case ".yaml", ".yml":
		return c.loadYAML(data)
	case ".ini":
		return c.loadINI(data)
	default:
		return errors.New("unsupported file extension")
	}
}

func (c *Configurable) loadJSON(data []byte) error {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return err
	}
	return c.setValuesFromMap(jsonData)
}

func (c *Configurable) loadYAML(data []byte) error {
	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return err
	}
	return c.setValuesFromMap(yamlData)
}

func (c *Configurable) loadINI(data []byte) error {
	cfg, err := ini.Load(data)
	if err != nil {
		return err
	}
	iniData := make(map[string]interface{})
	for key := range c.flags {
		if val := cfg.Section("").Key(key).String(); val != "" {
			iniData[key] = val
		}
	}
	return c.setValuesFromMap(iniData)
}

func (c *Configurable) setValuesFromMap(data map[string]interface{}) error {
	for key, value := range data {
		if flagVal, exists := c.flags[key]; exists {
			if err := c.setValue(flagVal, value); err != nil {
				return fmt.Errorf("error setting key %s: %w", key, err)
			}
		}
	}
	return nil
}

func (c *Configurable) setValue(flagVal interface{}, value interface{}) error {
	switch ptr := flagVal.(type) {
	case *int:
		intVal, err := toInt(value)
		if err != nil {
			return err
		}
		*ptr = intVal
	case *int64:
		int64Val, err := toInt64(value)
		if err != nil {
			return err
		}
		*ptr = int64Val
	case *float64:
		floatVal, err := toFloat64(value)
		if err != nil {
			return err
		}
		*ptr = floatVal
	case *string:
		strVal, err := toString(value)
		if err != nil {
			return err
		}
		*ptr = strVal
	case *bool:
		boolVal, err := toBool(value)
		if err != nil {
			return err
		}
		*ptr = boolVal
	case *time.Duration:
		strVal, err := toString(value)
		if err != nil {
			return err
		}
		duration, err := time.ParseDuration(strVal)
		if err != nil {
			return err
		}
		*ptr = duration
	case *ListFlag:
		listVal, err := toStringSlice(value)
		if err != nil {
			return err
		}
		*ptr.values = append(*ptr.values, listVal...)
	case *MapFlag:
		mapVal, err := toStringMap(value)
		if err != nil {
			return err
		}
		for k, v := range mapVal {
			(*ptr.values)[k] = v
		}
	default:
		return fmt.Errorf("unsupported flag type for key %v", ptr)
	}
	return nil
}

func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %v to int", value)
	}
}

func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %v to int64", value)
	}
}

func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %v to float64", value)
	}
}

func toString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("cannot convert %v to string", value)
	}
}

func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("cannot convert %v to bool", value)
	}
}

func toStringSlice(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case []interface{}:
		var result []string
		for _, item := range v {
			str, err := toString(item)
			if err != nil {
				return nil, err
			}
			result = append(result, str)
		}
		return result, nil
	case string:
		if v == "" {
			return []string{}, nil
		}
		return strings.Split(v, ","), nil
	default:
		return nil, fmt.Errorf("cannot convert %v to []string", value)
	}
}

func toStringMap(value interface{}) (map[string]string, error) {
	switch v := value.(type) {
	case map[string]interface{}:
		result := make(map[string]string)
		for key, val := range v {
			strVal, err := toString(val)
			if err != nil {
				return nil, err
			}
			result[key] = strVal
		}
		return result, nil
	case string:
		if v == "" {
			return map[string]string{}, nil
		}
		pairs := strings.Split(v, ",")
		result := make(map[string]string)
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("invalid map item: %s", pair)
			}
			result[kv[0]] = kv[1]
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %v to map[string]string", value)
	}
}

func (c *Configurable) checkAndSetFromEnv(name string) {
	if val, exists := os.LookupEnv(name); exists {
		if flagVal, exists := c.flags[name]; exists {
			c.setValue(flagVal, val)
		}
	}
}

func (c *Configurable) Usage() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Usage of %s:\n", os.Args[0])
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(&sb, "  -%s: %s (default: %s)\n", f.Name, f.Usage, f.DefValue)
	})
	return sb.String()
}

