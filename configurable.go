package configurable

import (
	`encoding/json`
	`errors`
	`flag`
	`fmt`
	`os`
	`path/filepath`
	`reflect`
	`strconv`
	`strings`
	`time`

	`github.com/go-ini/ini`
	`gopkg.in/yaml.v3`
)

type IConfigurable interface {
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

	LoadFile(filename string) error
	Parse(filename string) error

	Usage() string
}

type Configurable struct {
	flags map[string]interface{}
	err   error
}

func New() IConfigurable {
	return &Configurable{flags: make(map[string]interface{})}
}

func (c *Configurable) Int(name string) *int {
	c.checkAndSetFromEnv(name)
	val, _ := c.flags[name].(*int)
	return val
}

func (c *Configurable) NewInt(name string, value int, usage string) *int {
	var i = flag.Int(name, value, usage)
	c.flags[name] = i
	return i
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

func (c *Configurable) Parse(filename string) error {
	flag.Parse()
	if len(filename) == 0 {
		return nil
	}
	err := c.LoadFile(filename)
	if err != nil {
		return err
	}
	return nil
}

func (c *Configurable) Err() error {
	return c.err
}

func (c *Configurable) Value(name string) interface{} {
	return c.flags[name]
}

func (c *Configurable) LoadFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	ext := filepath.Ext(filepath.Base(filename))
	switch ext {
	case ".json":
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		if err != nil {
			return err
		}
		for key, value := range jsonData {
			if c.flags[key] != nil {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(value))
			}
		}
	case ".yaml", ".yml":
		var yamlData map[string]interface{}
		err = yaml.Unmarshal(data, &yamlData)
		if err != nil {
			return err
		}
		for key, value := range yamlData {
			if c.flags[key] != nil {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(value))
			}
		}
	case ".ini":
		cfg, err := ini.Load(data)
		if err != nil {
			return err
		}
		for key, _ := range c.flags {
			if cfg.Section("").HasKey(key) {
				reflect.ValueOf(c.flags[key]).Elem().Set(reflect.ValueOf(cfg.Section("").Key(key).Value()))
			}
		}
	default:
		return errors.New("unknown file type")
	}
	return nil
}

func (c *Configurable) checkAndSetFromEnv(name string) {
	if val, exists := os.LookupEnv(name); exists {
		// Check if the flag already exists
		if c.flags[name] != nil {
			switch v := c.flags[name].(type) {
			case *int:
				if parsedVal, err := strconv.Atoi(val); err == nil {
					*v = parsedVal
				}
			case *int64:
				if parsedVal, err := strconv.ParseInt(val, 10, 64); err == nil {
					*v = parsedVal
				}
			case *float32:
				if parsedVal, err := strconv.ParseFloat(val, 32); err == nil {
					*v = float32(parsedVal)
				}
			case *float64:
				if parsedVal, err := strconv.ParseFloat(val, 64); err == nil {
					*v = parsedVal
				}
			case *string:
				*v = val
			case *bool:
				if parsedVal, err := strconv.ParseBool(val); err == nil {
					*v = parsedVal
				}
			case *time.Duration:
				if parsedVal, err := time.ParseDuration(val); err == nil {
					*v = parsedVal
				}
			}
		}
	}
}

func (c *Configurable) Usage() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("%v [FLAGS]\n", os.Args[0]))
	builder.WriteString("Flag\tDefault\tDescription\tSource\n")

	flags := make([]*flag.Flag, 0)
	flag.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f)
	})

	nl, dl, ul, sl := 4, 7, 11, 6

	for _, f := range flags {
		source := "flag"
		if _, exists := os.LookupEnv(f.Name); exists {
			source = "env"
		} else if c.flags[f.Name] != nil {
			switch c.flags[f.Name].(type) {
			case *json.RawMessage:
				source = "json"
			case *yaml.Node:
				source = "yaml"
			case *ini.Key:
				source = "ini"
			}
		}

		builder.WriteString(fmt.Sprintf("-%-*s\t%-*s\t%-*s\t%s\n", nl, f.Name, dl, f.DefValue, ul, f.Usage, source))

		if len(f.Name)+1 > nl {
			nl = len(f.Name) + 1
		}
		if len(f.DefValue) > dl {
			dl = len(f.DefValue)
		}
		if len(f.Usage) > ul {
			ul = len(f.Usage)
		}
		if len(source) > sl {
			sl = len(source)
		}
	}

	builder.WriteString(fmt.Sprintf("%v\t%v\t%v\t%v\n", strings.Repeat("-", nl), strings.Repeat("-", dl), strings.Repeat("-", ul), strings.Repeat("-", sl)))

	return builder.String()
}
