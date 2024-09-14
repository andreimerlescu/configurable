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
    err   error
}

func New() IConfigurable {
    return &Configurable{flags: make(map[string]interface{})}
}

// Existing methods for Int, Int64, Float64, String, Bool, Duration...

// ListFlag implements flag.Value for []string
type ListFlag []string

func (l *ListFlag) String() string {
    return strings.Join(*l, ",")
}

func (l *ListFlag) Set(value string) error {
    items := strings.Split(value, ",")
    *l = append(*l, items...)
    return nil
}

func (c *Configurable) List(name string) *[]string {
    c.checkAndSetFromEnv(name)
    if val, ok := c.flags[name].(*ListFlag); ok {
        return (*[]string)(val)
    }
    return nil
}

func (c *Configurable) NewList(name string, value []string, usage string) *[]string {
    l := ListFlag(value)
    flag.Var(&l, name, usage)
    c.flags[name] = &l
    return (*[]string)(&l)
}

// MapFlag implements flag.Value for map[string]string
type MapFlag map[string]string

func (m *MapFlag) String() string {
    var entries []string
    for k, v := range *m {
        entries = append(entries, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(entries, ",")
}

func (m *MapFlag) Set(value string) error {
    pairs := strings.Split(value, ",")
    for _, pair := range pairs {
        kv := strings.SplitN(pair, "=", 2)
        if len(kv) != 2 {
            return fmt.Errorf("invalid map item: %s", pair)
        }
        (*m)[kv[0]] = kv[1]
    }
    return nil
}

func (c *Configurable) Map(name string) *map[string]string {
    c.checkAndSetFromEnv(name)
    if val, ok := c.flags[name].(*MapFlag); ok {
        return (*map[string]string)(val)
    }
    return nil
}

func (c *Configurable) NewMap(name string, value map[string]string, usage string) *map[string]string {
    m := MapFlag(value)
    flag.Var(&m, name, usage)
    c.flags[name] = &m
    return (*map[string]string)(&m)
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
                switch v := c.flags[key].(type) {
                case *int:
                    if num, ok := value.(float64); ok {
                        *v = int(num)
                    }
                case *int64:
                    if num, ok := value.(float64); ok {
                        *v = int64(num)
                    }
                case *float64:
                    if num, ok := value.(float64); ok {
                        *v = num
                    }
                case *string:
                    if str, ok := value.(string); ok {
                        *v = str
                    }
                case *bool:
                    if b, ok := value.(bool); ok {
                        *v = b
                    }
                case *time.Duration:
                    if str, ok := value.(string); ok {
                        if parsedVal, err := time.ParseDuration(str); err == nil {
                            *v = parsedVal
                        }
                    }
                case *ListFlag:
                    if arr, ok := value.([]interface{}); ok {
                        for _, item := range arr {
                            if str, ok := item.(string); ok {
                                v.Set(str)
                            }
                        }
                    }
                case *MapFlag:
                    if m, ok := value.(map[string]interface{}); ok {
                        for mk, mv := range m {
                            if str, ok := mv.(string); ok {
                                v.Set(fmt.Sprintf("%s=%s", mk, str))
                            }
                        }
                    }
                }
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
                switch v := c.flags[key].(type) {
                case *int:
                    if num, ok := value.(int); ok {
                        *v = num
                    } else if num, ok := value.(float64); ok {
                        *v = int(num)
                    }
                case *int64:
                    if num, ok := value.(int64); ok {
                        *v = num
                    } else if num, ok := value.(float64); ok {
                        *v = int64(num)
                    }
                case *float64:
                    if num, ok := value.(float64); ok {
                        *v = num
                    }
                case *string:
                    if str, ok := value.(string); ok {
                        *v = str
                    }
                case *bool:
                    if b, ok := value.(bool); ok {
                        *v = b
                    }
                case *time.Duration:
                    if str, ok := value.(string); ok {
                        if parsedVal, err := time.ParseDuration(str); err == nil {
                            *v = parsedVal
                        }
                    }
                case *ListFlag:
                    if arr, ok := value.([]interface{}); ok {
                        for _, item := range arr {
                            if str, ok := item.(string); ok {
                                v.Set(str)
                            }
                        }
                    }
                case *MapFlag:
                    if m, ok := value.(map[string]interface{}); ok {
                        for mk, mv := range m {
                            if str, ok := mv.(string); ok {
                                v.Set(fmt.Sprintf("%s=%s", mk, str))
                            }
                        }
                    }
                }
            }
        }
    case ".ini":
        cfg, err := ini.Load(data)
        if err != nil {
            return err
        }
        for key := range c.flags {
            if cfg.Section("").HasKey(key) {
                val := cfg.Section("").Key(key).String()
                switch v := c.flags[key].(type) {
                case *string:
                    *v = val
                case *int:
                    if parsedVal, err := strconv.Atoi(val); err == nil {
                        *v = parsedVal
                    }
                case *bool:
                    if parsedVal, err := strconv.ParseBool(val); err == nil {
                        *v = parsedVal
                    }
                case *ListFlag:
                    v.Set(val)
                case *MapFlag:
                    v.Set(val)
                }
            }
        }
    default:
        return errors.New("unknown file type")
    }
    return nil
}

func (c *Configurable) checkAndSetFromEnv(name string) {
    if val, exists := os.LookupEnv(name); exists {
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
            case *ListFlag:
                v.Set(val)
            case *MapFlag:
                v.Set(val)
            }
        }
    }
}

func (c *Configurable) Usage() string {
    var builder strings.Builder

    builder.WriteString(fmt.Sprintf("%v [FLAGS]\n", os.Args[0]))

    flags := make([]*flag.Flag, 0)
    flag.VisitAll(func(f *flag.Flag) {
        flags = append(flags, f)
    })

    nl, dl, ul, sl := 4, 7, 11, 6

    for _, f := range flags {
        source := "flag"
        if _, exists := os.LookupEnv(f.Name); exists {
            source = "env"
        }
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

    builder.WriteString(
        fmt.Sprintf("Flag%v\tDefault%v\tDescription%v\tSource%v\n",
            strings.Repeat(" ", min_zero(nl-4)),
            strings.Repeat(" ", min_zero(dl-7)),
            strings.Repeat(" ", min_zero(ul-11)),
            strings.Repeat(" ", min_zero(sl-6))))
    builder.WriteString(
        fmt.Sprintf("%v\t%v\t%v\t%v\n",
            strings.Repeat("-", min_zero(nl)),
            strings.Repeat("-", min_zero(dl)),
            strings.Repeat("-", min_zero(ul)),
            strings.Repeat("-", min_zero(sl))))

    for _, f := range flags {
        source := "flag"
        if _, exists := os.LookupEnv(f.Name); exists {
            source = "env"
        }
        builder.WriteString(fmt.Sprintf("-%-*s\t%-*s\t%-*s\t%s\n", nl, f.Name, dl, f.DefValue, ul, f.Usage, source))
    }

    return builder.String()
}

func min_zero(number int) int {
    if number < 0 {
        return 0
    } else {
        return number
    }
}