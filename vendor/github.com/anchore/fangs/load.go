package fangs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func Load(cfg Config, cmd *cobra.Command, configurations ...any) error {
	return loadConfig(cfg, commandFlagRefs(cmd), configurations...)
}

func LoadAt(cfg Config, cmd *cobra.Command, path string, configuration any) error {
	t := reflect.TypeOf(configuration)
	config := reflect.StructOf([]reflect.StructField{{
		Name: upperFirst(path),
		Type: t,
		Tag:  reflect.StructTag(fmt.Sprintf(`%s:"%s"`, cfg.TagName, path)),
	}})

	value := reflect.New(config)
	value.Elem().Field(0).Set(reflect.ValueOf(configuration))

	return Load(cfg, cmd, value.Interface())
}

func loadConfig(cfg Config, flags flagRefs, configurations ...any) error {
	// ensure the config is set up sufficiently
	if cfg.Logger == nil || cfg.Finders == nil {
		return fmt.Errorf("config.Load requires logger and finders to be set, but only has %+v", cfg)
	}

	// allow for nested options to be specified via environment variables
	// e.g. pod.context = APPNAME_POD_CONTEXT
	v := viper.NewWithOptions(viper.EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")))

	for _, configuration := range configurations {
		if !isPtr(reflect.TypeOf(configuration)) {
			return fmt.Errorf("config.Load configuration parameters must be a pointers, got: %s -- %v", reflect.TypeOf(configuration).Name(), configuration)
		}
	}

	// priority order: viper.Set, flag, env, config, kv, defaults
	// flags have already been loaded into viper by command construction

	// check if user specified config; otherwise read all possible paths
	if err := readConfigFile(cfg, v); err != nil {
		if isNotFoundErr(err) {
			cfg.Logger.Debug("no config file found, using defaults")
		} else {
			return fmt.Errorf("unable to load config: %w", err)
		}
	}

	// load environment variables
	v.SetEnvPrefix(cfg.AppName)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	for _, configuration := range configurations {
		configureViper(cfg, v, reflect.ValueOf(configuration), flags, []string{})

		// unmarshal fully populated viper object onto config
		err := v.Unmarshal(configuration, func(dc *mapstructure.DecoderConfig) {
			dc.TagName = cfg.TagName
			// ZeroFields will use what is present in the config file instead of modifying existing defaults
			dc.ZeroFields = true
		})
		if err != nil {
			return err
		}

		// Convert all populated config options to their internal application values ex: scope string => scopeOpt source.Scope
		err = postLoad(reflect.ValueOf(configuration))
		if err != nil {
			return err
		}
	}

	return nil
}

// configureViper loads the default configuration values into the viper instance,
// before the config values are read and parsed. the value _must_ be a pointer but
// may be a pointer to a pointer
func configureViper(cfg Config, vpr *viper.Viper, v reflect.Value, flags flagRefs, path []string) {
	t := v.Type()
	if !isPtr(t) {
		panic(fmt.Sprintf("configureViper v must be a pointer, got: %#v", v))
	}

	// v is always a pointer
	ptr := v.Pointer()
	t = t.Elem()
	v = v.Elem()

	// might be a pointer value
	for isPtr(t) {
		t = t.Elem()
		v = v.Elem()
	}

	if !isStruct(t) {
		envVar := envVar(cfg.AppName, path...)
		path := strings.Join(path, ".")

		if flag, ok := flags[ptr]; ok {
			cfg.Logger.Tracef("binding env var w/flag: %s", envVar)
			err := vpr.BindPFlag(path, flag)
			if err != nil {
				cfg.Logger.Debugf("unable to bind flag: %s to %#v", path, flag)
			}
			return
		}

		cfg.Logger.Tracef("binding env var: %s", envVar)

		vpr.SetDefault(path, nil) // no default value actually needs to be set for Viper to read config values
		return
	}

	// for each field in the configuration struct, see if the field implements the defaultValueLoader interface and invoke it if it does
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		path := path
		if tag, ok := f.Tag.Lookup(cfg.TagName); ok {
			// handle ,squash mapstructure tags
			parts := strings.Split(tag, ",")
			tag = parts[0]
			if tag == "-" {
				continue
			}
			switch {
			case contains(parts, "squash"):
				// use the current path
			case tag == "":
				path = append(path, f.Name)
			default:
				path = append(path, tag)
			}
		} else {
			path = append(path, f.Name)
		}

		v := v.Field(i)

		t := f.Type
		if isPtr(t) && v.IsNil() {
			t = t.Elem()
			if isStruct(t) {
				newV := reflect.New(t)
				v.Set(newV)
			}
		}

		configureViper(cfg, vpr, v.Addr(), flags, path)
	}
}

func readConfigFile(cfg Config, v *viper.Viper) error {
	for _, finder := range cfg.Finders {
		for _, file := range finder(cfg) {
			if !fileExists(file) {
				continue
			}
			v.SetConfigFile(file)
			err := v.ReadInConfig()
			if isNotFoundErr(err) {
				continue
			}
			if err != nil {
				return err
			}
			v.Set("config", v.ConfigFileUsed())
			return nil
		}
	}
	return &viper.ConfigFileNotFoundError{}
}

func postLoad(v reflect.Value) error {
	t := v.Type()

	for isPtr(t) {
		if v.IsNil() {
			return nil
		}

		obj := v.Interface()
		if p, ok := obj.(PostLoader); ok && !isPromotedMethod(obj, "PostLoad") {
			if err := p.PostLoad(); err != nil {
				return err
			}
		}

		t = t.Elem()
		v = v.Elem()
	}

	switch {
	case isStruct(t):
		return postLoadStruct(v)
	case isSlice(t):
		return postLoadSlice(v)
	case isMap(t):
		return postLoadMap(v)
	}

	return nil
}

// postLoadStruct call recursively on struct fields
func postLoadStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		v := v.Field(i)

		if isNil(v) {
			continue
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			continue
		}

		if err := postLoad(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}

// postLoadSlice call recursively on slice items
func postLoadSlice(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		v := v.Index(i)

		if isNil(v) {
			continue
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			continue
		}

		if err := postLoad(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}

// postLoadMap call recursively on map values
func postLoadMap(v reflect.Value) error {
	mapV := v
	i := v.MapRange()
	for i.Next() {
		v := i.Value()

		if isNil(v) {
			continue
		}

		for isPtr(v.Type()) {
			v = v.Elem()
		}

		if !v.CanAddr() {
			// unable to call .Addr() on struct map entries, so copy to a new instance and set on the map
			if isStruct(v.Type()) {
				newV := reflect.New(v.Type())
				newV.Elem().Set(v)
				if err := postLoad(newV); err != nil {
					return err
				}
				mapV.SetMapIndex(i.Key(), newV.Elem())
			}

			continue
		}

		if err := postLoad(v.Addr()); err != nil {
			return err
		}
	}
	return nil
}

type flagRefs map[uintptr]*pflag.Flag

func commandFlagRefs(cmd *cobra.Command) flagRefs {
	return getFlagRefs(cmd.PersistentFlags(), cmd.Flags())
}

func getFlagRefs(flagSets ...*pflag.FlagSet) flagRefs {
	refs := flagRefs{}
	for _, flags := range flagSets {
		flags.VisitAll(func(flag *pflag.Flag) {
			refs[getFlagRef(flag)] = flag
		})
	}
	return refs
}

func getFlagRef(flag *pflag.Flag) uintptr {
	v := reflect.ValueOf(flag.Value)

	// check for struct types like stringArrayValue
	if isPtr(v.Type()) {
		vf := v.Elem()
		vt := vf.Type()
		if isStruct(vt) {
			if _, ok := vt.FieldByName("value"); ok {
				vf = vf.FieldByName("value")
				if vf.IsValid() {
					v = vf
				}
			}
		}
	}
	return v.Pointer()
}

func upperFirst(p string) string {
	if len(p) < 2 {
		return strings.ToUpper(p)
	}
	return strings.ToUpper(p[0:1]) + p[1:]
}

func isPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr
}

func isStruct(typ reflect.Type) bool {
	return typ.Kind() == reflect.Struct
}

func isSlice(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice
}

func isMap(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map
}

func isNil(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Type().Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

func isNotFoundErr(err error) bool {
	var notFound *viper.ConfigFileNotFoundError
	return err != nil && errors.As(err, &notFound)
}
