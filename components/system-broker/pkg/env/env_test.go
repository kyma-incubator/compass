package env_test

import (
	"context"
	"fmt"
	"github.com/fatih/structs"
	"github.com/kyma-incubator/compass/components/system-broker/internal/config"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	mapKey           = "mapkey"
	key              = "key"
	description      = "desc"
	flagDefaultValue = "pflagDefaultValue"
	fileValue        = "fileValue"
	envValue         = "envValue"
	flagValue        = "pflagValue"
	overrideValue    = "overrideValue"

	keyWbool      = "wbool"
	keyWint       = "wint"
	keyWstring    = "wstring"
	keyWmappedVal = "w_mapped_val"

	keyNbool      = "nest.nbool"
	keyNint       = "nest.nint"
	keyNstring    = "nest.nstring"
	keyNslice     = "nest.nslice"
	keyNmappedVal = "nest.n_mapped_val"

	keySquashNbool      = "nbool"
	keySquashNint       = "nint"
	keySquashNstring    = "nstring"
	keySquashNslice     = "nslice"
	keySquashNmappedVal = "n_mapped_val"

	keyMapNbool      = "wmapnest" + "." + mapKey + "." + "nbool"
	keyMapNint       = "wmapnest" + "." + mapKey + "." + "nint"
	keyMapNstring    = "wmapnest" + "." + mapKey + "." + "nstring"
	keyMapNslice     = "wmapnest" + "." + mapKey + "." + "nslice"
	keyMapNmappedVal = "wmapnest" + "." + mapKey + "." + "n_mapped_val"

	keyLogFormat = "log.format"
	keyLogLevel  = "log.level"
	keyOutput = "log.output"
	keyBootstrapCorrelationID = "log.bootstrap_correlation_id"

	keyFileName     = "file.name"
	keyFileLocation = "file.location"
	keyFileFormat   = "file.format"
)

type Nest struct {
	NBool      bool
	NInt       int
	NString    string
	NSlice     []string
	NMappedVal string `mapstructure:"n_mapped_val" structs:"n_mapped_val"  yaml:"n_mapped_val"`
}

type Outer struct {
	WBool      bool
	WInt       int
	WString    string
	WMappedVal string `mapstructure:"w_mapped_val" structs:"w_mapped_val" yaml:"w_mapped_val"`
	WMapNest   map[string]Nest
	Nest       Nest
	Squash     Nest `mapstructure:",squash"`
	Log        *log.Config
}

type FlatOuter struct {
	WBool      bool
	WInt       int
	WString    string
	WMappedVal string `mapstructure:"w_mapped_val" structs:"w_mapped_val" yaml:"w_mapped_val"`
	WMapNest   map[string]Nest
	Nest       Nest

	// Flattened Nest fields due to squash tag
	NBool      bool
	NInt       int
	NString    string
	NSlice     []string
	NMappedVal string `mapstructure:"n_mapped_val" structs:"n_mapped_val" yaml:"n_mapped_val"`

	Log *log.Config
}

type testFile struct {
	env.File
	content interface{}
}

func TestEnvSuite(t *testing.T) {
	suite.Run(t, new(EnvSuite))
}

type EnvSuite struct {
	suite.Suite

	outer     Outer
	flatOuter FlatOuter

	cfgFile   testFile
	testFlags *pflag.FlagSet

	environment env.Environment
	err         error
}

// BeforeEach
func (suite *EnvSuite) SetupTest() {
	suite.testFlags = env.EmptyFlagSet()

	nest := Nest{
		NBool:      true,
		NInt:       4321,
		NString:    "nstringval",
		NSlice:     []string{"nval1", "nval2", "nval3"},
		NMappedVal: "nmappedval",
	}

	suite.outer = Outer{
		WBool:      true,
		WInt:       1234,
		WString:    "wstringval",
		WMappedVal: "wmappedval",
		Squash:     nest,
		Log:        log.DefaultConfig(),
		Nest:       nest,
		WMapNest: map[string]Nest{
			mapKey: nest,
		},
	}

	suite.flatOuter = FlatOuter{
		WBool:      true,
		WInt:       1234,
		WString:    "wstringval",
		WMappedVal: "wmappedval",
		NBool:      true,
		NInt:       4321,
		NString:    "nstringval",
		NSlice:     []string{"nval1", "nval2", "nval3"},
		NMappedVal: "nmappedval",
		Log:        log.DefaultConfig(),
		Nest:       nest,
		WMapNest: map[string]Nest{
			mapKey: nest,
		},
	}
}

// AfterEach
func (suite *EnvSuite) TearDownTest() {
	suite.cleanUpEnvVars()
	suite.cleanUpFlags()
	suite.cleanUpFile()
}

// Tests

func (suite *EnvSuite) TestNewAddsViperBindingsForTheProvidedFlags() {
	suite.testFlags.AddFlagSet(standardPFlagsSet(suite.outer))
	suite.cfgFile.content = nil

	suite.verifyEnvCreated()

	suite.verifyEnvContainsValues(suite.flatOuter)
}

func (suite *EnvSuite) TestNewDoesNotFailWhenConfigFileDoesNotExists() {
	_, err := env.New(context.TODO(), suite.testFlags)
	suite.Require().NoError(err)
}

func (suite *EnvSuite) TestNewWhenConfigFileFlagsAreNotProvidedShouldNotFail() {
	suite.cfgFile = testFile{
		File:    env.DefaultConfigFile(),
		content: suite.flatOuter,
	}
	defer suite.cleanUpFile()

	suite.Require().Nil(suite.testFlags.Lookup(keyFileName))
	suite.Require().Nil(suite.testFlags.Lookup(keyFileLocation))
	suite.Require().Nil(suite.testFlags.Lookup(keyFileFormat))

	suite.verifyEnvCreated()

	suite.Require().Nil(suite.environment.Get(keyFileName))
	suite.Require().Nil(suite.environment.Get(keyFileName))
	suite.Require().Nil(suite.environment.Get(keyFileName))
}

func (suite *EnvSuite) TestNewWhenConfigFileFlagsAreProvided() {
	var tests = []struct {
		Msg      string
		TestFunc func()
	}{
		{
			Msg: "allows obtaining config file values from the environment",
			TestFunc: func() {
				suite.verifyEnvCreated()
				suite.verifyEnvContainsValues(struct{ File env.File }{File: suite.cfgFile.File})
			},
		},
		{
			Msg: "allows unmarshaling config file values from the environment",
			TestFunc: func() {
				suite.verifyEnvCreated()

				file := testFile{}
				suite.Require().NoError(suite.environment.Unmarshal(&file))
				suite.Require().Equal(file.File, suite.cfgFile.File)
			},
		},
		{
			Msg: "allows overriding the config file properties",
			TestFunc: func() {
				suite.cfgFile.Name = "updatedName"
				suite.Require().NoError(suite.testFlags.Set(keyFileName, "updatedName"))
				suite.verifyEnvCreated()

				suite.verifyEnvContainsValues(struct{ File env.File }{File: suite.cfgFile.File})
			},
		},
		{
			Msg: "reads the file in the environment",
			TestFunc: func() {
				suite.verifyEnvCreated()
				suite.verifyEnvContainsValues(suite.flatOuter)
			},
		},
		{
			Msg: "returns an err if config file loading fails",
			TestFunc: func() {
				suite.cfgFile.Format = "json"
				suite.testFlags.Set(keyFileFormat, "json")

				suite.Require().Error(suite.createEnv())
			},
		},
		{
			Msg: "when the logging properties are changed reconfigures the loggers with the correct logging config",
			TestFunc: func() {
				suite.verifyEnvCreated()
				oldCfg := log.Configuration()
				newLogLevel := logrus.DebugLevel.String()
				suite.Require().NotEqual(newLogLevel, oldCfg.Level)
				suite.Require().NotEqual(log.D().Logger.Level.String(), newLogLevel)
				newOutput := os.Stderr.Name()
				suite.Require().NotEqual(newOutput, oldCfg.Output)
				suite.Require().NotEqual(log.D().Logger.Out.(*os.File).Name(), newOutput)

				f := suite.cfgFile.Location + string(filepath.Separator) + suite.cfgFile.Name + "." + suite.cfgFile.Format
				fileContent := suite.cfgFile.content.(FlatOuter)
				fileContent.Log = log.DefaultConfig()
				fileContent.Log.Level = logrus.DebugLevel.String()
				fileContent.Log.Output = newOutput
				suite.cfgFile.content = fileContent
				bytes, err := yaml.Marshal(suite.cfgFile.content)
				yamlFile := strings.ReplaceAll(string(bytes), "bootstrapcorrelationid", "bootstrap_correlation_id")
				suite.Require().NoError(err)
				err = ioutil.WriteFile(f, []byte(yamlFile), 0640)
				suite.Require().NoError(err)

				suite.Require().Eventually(func() bool {
					return log.D().Logger.IsLevelEnabled(logrus.DebugLevel)
				}, time.Minute, 5*time.Millisecond)
				suite.Require().Equal(log.Configuration().Level, newLogLevel)
				suite.Require().Equal(log.Configuration().Output, newOutput)
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.Msg, func() {
			defer suite.TearDownTest()
			suite.cfgFile = testFile{
				File:    env.DefaultConfigFile(),
				content: suite.flatOuter,
			}
			config.AddPFlags(suite.testFlags)

			test.TestFunc()
		})
	}
}

func (suite *EnvSuite) TestBindPFlagAllowsGettingAFlagFromTheEnvironmentWithAnAliasName() {
	const (
		key         = "test_flag"
		aliasKey    = "test.flag"
	)

	suite.testFlags.AddFlagSet(singlePFlagSet(key, flagDefaultValue, description))

	suite.verifyEnvCreated()

	suite.Require().Equal(suite.environment.Get(key),flagDefaultValue)
	suite.Require().Nil(suite.environment.Get(aliasKey))

	err := suite.environment.BindPFlag(aliasKey, suite.testFlags.Lookup(key))
	suite.Require().NoError(err)

	suite.Require().Equal(suite.environment.Get(key),flagDefaultValue)
	suite.Require().Equal(suite.environment.Get(aliasKey), flagDefaultValue)
}

// Helper functions

func generatedPFlagsSet(s interface{}) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	env.CreatePFlags(set, s)

	return set
}

func standardPFlagsSet(s Outer) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)

	set.Bool(keyWbool, s.WBool, description)
	set.Int(keyWint, s.WInt, description)
	set.String(keyWstring, s.WString, description)
	set.String(keyWmappedVal, s.WMappedVal, description)

	set.Bool(keySquashNbool, s.Squash.NBool, description)
	set.Int(keySquashNint, s.Squash.NInt, description)
	set.String(keySquashNstring, s.Squash.NString, description)
	set.StringSlice(keySquashNslice, s.Squash.NSlice, description)
	set.String(keySquashNmappedVal, s.Squash.NMappedVal, description)

	set.Bool(keyNbool, s.Nest.NBool, description)
	set.Int(keyNint, s.Nest.NInt, description)
	set.String(keyNstring, s.Nest.NString, description)
	set.StringSlice(keyNslice, s.Nest.NSlice, description)
	set.String(keyNmappedVal, s.Nest.NMappedVal, description)

	set.Bool(keyMapNbool, s.WMapNest[mapKey].NBool, description)
	set.Int(keyMapNint, s.WMapNest[mapKey].NInt, description)
	set.String(keyMapNstring, s.WMapNest[mapKey].NString, description)
	set.StringSlice(keyMapNslice, s.WMapNest[mapKey].NSlice, description)
	set.String(keyMapNmappedVal, s.WMapNest[mapKey].NMappedVal, description)

	set.String(keyLogLevel, s.Log.Level, description)
	set.String(keyLogFormat, s.Log.Format, description)
	set.String(keyOutput, s.Log.Output, description)
	set.String(keyBootstrapCorrelationID, s.Log.BootstrapCorrelationID, description)

	return set
}

func singlePFlagSet(key, defaultValue, description string) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	set.String(key, defaultValue, description)

	return set
}

func (suite *EnvSuite) setPFlags(o Outer) {
	suite.Require().NoError(suite.testFlags.Set(keyWbool, cast.ToString(o.WBool)))
	suite.Require().NoError(suite.testFlags.Set(keyWint, cast.ToString(o.WInt)))
	suite.Require().NoError(suite.testFlags.Set(keyWstring, o.WString))
	suite.Require().NoError(suite.testFlags.Set(keyWmappedVal, o.WMappedVal))

	suite.Require().NoError(suite.testFlags.Set(keySquashNbool, cast.ToString(o.Squash.NBool)))
	suite.Require().NoError(suite.testFlags.Set(keySquashNint, cast.ToString(o.Squash.NInt)))
	suite.Require().NoError(suite.testFlags.Set(keySquashNstring, o.Squash.NString))
	suite.Require().NoError(suite.testFlags.Set(keySquashNslice, strings.Join(o.Squash.NSlice, ",")))
	suite.Require().NoError(suite.testFlags.Set(keySquashNmappedVal, o.Squash.NMappedVal))

	suite.Require().NoError(suite.testFlags.Set(keyNbool, cast.ToString(o.Nest.NBool)))
	suite.Require().NoError(suite.testFlags.Set(keyNint, cast.ToString(o.Nest.NInt)))
	suite.Require().NoError(suite.testFlags.Set(keyNstring, o.Nest.NString))
	suite.Require().NoError(suite.testFlags.Set(keyNmappedVal, o.Nest.NMappedVal))

	suite.Require().NoError(suite.testFlags.Set(keyMapNbool, cast.ToString(o.WMapNest[mapKey].NBool)))
	suite.Require().NoError(suite.testFlags.Set(keyMapNint, cast.ToString(o.WMapNest[mapKey].NInt)))
	suite.Require().NoError(suite.testFlags.Set(keyMapNstring, o.WMapNest[mapKey].NString))
	suite.Require().NoError(suite.testFlags.Set(keyMapNmappedVal, o.WMapNest[mapKey].NMappedVal))

	suite.Require().NoError(suite.testFlags.Set(keyLogFormat, o.Log.Format))
	suite.Require().NoError(suite.testFlags.Set(keyLogLevel, o.Log.Level))
	suite.Require().NoError(suite.testFlags.Set(keyOutput, o.Log.Output))
	suite.Require().NoError(suite.testFlags.Set(keyBootstrapCorrelationID, o.Log.BootstrapCorrelationID))
}

func (suite *EnvSuite) setEnvVars() {
	suite.Require().NoError(os.Setenv(strings.ToTitle(keyWbool), cast.ToString(suite.outer.WBool)))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keyWint), cast.ToString(suite.outer.WInt)))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keyWstring), suite.outer.WString))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keyWmappedVal), suite.outer.WMappedVal))

	suite.Require().NoError(os.Setenv(strings.ToTitle(keySquashNbool), cast.ToString(suite.outer.Squash.NBool)))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keySquashNint), cast.ToString(suite.outer.Squash.NInt)))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keySquashNstring), suite.outer.Squash.NString))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keySquashNslice), strings.Join(suite.outer.Squash.NSlice, ",")))
	suite.Require().NoError(os.Setenv(strings.ToTitle(keySquashNmappedVal), suite.outer.Squash.NMappedVal))

	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyNbool), ".", "_", -1), cast.ToString(suite.outer.Nest.NBool)))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyNint), ".", "_", -1), cast.ToString(suite.outer.Nest.NInt)))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyNstring), ".", "_", -1), suite.outer.Nest.NString))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyNslice), ".", "_", -1), strings.Join(suite.outer.Nest.NSlice, ",")))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyNmappedVal), ".", "_", -1), suite.outer.Nest.NMappedVal))

	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyMapNbool), ".", "_", -1), cast.ToString(suite.outer.WMapNest[mapKey].NBool)))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyMapNint), ".", "_", -1), cast.ToString(suite.outer.WMapNest[mapKey].NInt)))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyMapNstring), ".", "_", -1), suite.outer.WMapNest[mapKey].NString))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyMapNslice), ".", "_", -1), strings.Join(suite.outer.WMapNest[mapKey].NSlice, ",")))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyMapNmappedVal), ".", "_", -1), suite.outer.WMapNest[mapKey].NMappedVal))

	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyLogFormat), ".", "_", -1), suite.outer.Log.Format))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyLogLevel), ".", "_", -1), suite.outer.Log.Level))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyOutput), ".", "_", -1), suite.outer.Log.Output))
	suite.Require().NoError(os.Setenv(strings.Replace(strings.ToTitle(keyBootstrapCorrelationID), ".", "_", -1), suite.outer.Log.BootstrapCorrelationID))
}

func (suite *EnvSuite) cleanUpEnvVars() {
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keyWbool)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keyWint)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keyWstring)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keyWmappedVal)))

	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keySquashNbool)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keySquashNint)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keySquashNstring)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keySquashNslice)))
	suite.Require().NoError(os.Unsetenv(strings.ToTitle(keySquashNmappedVal)))

	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyNbool), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyNint), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyNstring), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyNslice), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyNmappedVal), ".", "_", -1)))

	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyMapNbool), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyMapNint), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyMapNstring), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyMapNslice), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyMapNmappedVal), ".", "_", -1)))

	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyLogFormat), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyLogLevel), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyOutput), ".", "_", -1)))
	suite.Require().NoError(os.Unsetenv(strings.Replace(strings.ToTitle(keyBootstrapCorrelationID), ".", "_", -1)))

	suite.Require().NoError(os.Unsetenv(strings.ToTitle(key)))
}

func (suite *EnvSuite) cleanUpFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	suite.testFlags = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
}

func (suite *EnvSuite) createEnv() error {
	if suite.cfgFile.content != nil {
		f := suite.cfgFile.Location + string(filepath.Separator) + suite.cfgFile.Name + "." + suite.cfgFile.Format
		bytes, err := yaml.Marshal(suite.cfgFile.content)
		suite.Require().NoError(err)
		err = ioutil.WriteFile(f, bytes, 0640)
		suite.Require().NoError(err)
	}

	suite.environment, suite.err = env.Default(context.TODO(), func(set *pflag.FlagSet) {
		set.AddFlagSet(suite.testFlags)
	})
	return suite.err
}

func (suite *EnvSuite) cleanUpFile() {
	f := suite.cfgFile.Location + string(filepath.Separator) + suite.cfgFile.Name + "." + suite.cfgFile.Format
	os.Remove(f)
}

func (suite *EnvSuite) verifyEnvCreated() {
	suite.Require().NoError(suite.createEnv())
}

func (suite *EnvSuite) verifyValues(fields map[string]interface{}, prefix string) {
	for name, value := range fields {
		switch v := value.(type) {
		case map[string]interface{}:
			suite.verifyValues(v, prefix+name+".")
		case []string:
			switch envVar := suite.environment.Get(prefix + name).(type) {
			case string:
				suite.Require().Equal(envVar, strings.Join(v, ","))
			case []string, []interface{}:
				suite.Require().Equal(fmt.Sprint(envVar), fmt.Sprint(v))
			default:
				suite.T().Fatalf("Expected env value of type []string but got: %T", envVar)
			}
		default:
			suite.Require().Equal(cast.ToString(suite.environment.Get(prefix+name)), cast.ToString(v), prefix+name)
		}
	}
}

func (suite *EnvSuite) verifyEnvContainsValues(expected interface{}) {
	fields := structs.Map(expected)
	suite.verifyValues(fields, "")
}
