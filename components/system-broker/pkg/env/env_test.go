package env_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/internal/config"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"

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
	Squash     Nest `mapstructure:",squash" structs:"squash,flatten"`
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
	suite.testFlags.AddFlagSet(suite.standardPFlagsSet(suite.outer))
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
				err := suite.testFlags.Set(keyFileFormat, "json")

				suite.Require().NoError(err)
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
				fileContent, ok := suite.cfgFile.content.(FlatOuter)
				suite.Require().True(ok)
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
		key      = "test_flag"
		aliasKey = "test.flag"
	)

	suite.testFlags.AddFlagSet(singlePFlagSet(key, flagDefaultValue, description))

	suite.verifyEnvCreated()

	suite.Require().Equal(suite.environment.Get(key), flagDefaultValue)
	suite.Require().Nil(suite.environment.Get(aliasKey))

	err := suite.environment.BindPFlag(aliasKey, suite.testFlags.Lookup(key))
	suite.Require().NoError(err)

	suite.Require().Equal(suite.environment.Get(key), flagDefaultValue)
	suite.Require().Equal(suite.environment.Get(aliasKey), flagDefaultValue)
}

func (suite *EnvSuite) TestGetWhenPropertiesAreLoadedViaPFlags() {
	var overrideOuter Outer
	var overrideOuterOutput FlatOuter

	ovrNest := Nest{
		NBool:      false,
		NInt:       9999,
		NString:    "overrideval",
		NSlice:     []string{"nval1", "nval2", "nval3"},
		NMappedVal: "overrideval",
	}

	overrideOuter = Outer{
		WBool:      false,
		WInt:       8888,
		WString:    "overrideval",
		WMappedVal: "overrideval",
		Nest:       ovrNest,
		Squash:     ovrNest,
		Log:        log.DefaultConfig(),
	}

	overrideOuterOutput = FlatOuter{
		WBool:      false,
		WInt:       8888,
		WString:    "overrideval",
		WMappedVal: "overrideval",
		Nest:       ovrNest,
		NBool:      false,
		NInt:       9999,
		NString:    "overrideval",
		NSlice:     []string{"nval1", "nval2", "nval3"},
		NMappedVal: "overrideval",
		Log:        log.DefaultConfig(),
	}

	var tests = []struct {
		Msg       string
		SetUpFunc func()
		TestFunc  func()
	}{
		{
			Msg: "standard pflags returns the default flag value if the flag is not set",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(suite.standardPFlagsSet(suite.outer))
			},
			TestFunc: func() {
				suite.verifyEnvContainsValues(suite.flatOuter)
			},
		},
		{
			Msg: "standard pflags returns the flags values if the flags are set",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(suite.standardPFlagsSet(suite.outer))
			},
			TestFunc: func() {
				suite.setPFlags(overrideOuter)
				suite.verifyEnvContainsValues(overrideOuterOutput)
			},
		},
		{
			Msg: "generated pflags returns the default flag value if the flag is not set",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(generatedPFlagsSet(suite.outer))
			},
			TestFunc: func() {
				suite.verifyEnvContainsValues(suite.flatOuter)
			},
		},
		{
			Msg: "generated pflags returns the flags values if the flags are set",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(generatedPFlagsSet(suite.outer))
			},
			TestFunc: func() {
				suite.setPFlags(overrideOuter)
				suite.verifyEnvContainsValues(overrideOuterOutput)
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.Msg, func() {
			defer suite.TearDownTest()
			test.SetUpFunc()

			suite.verifyEnvCreated()

			test.TestFunc()
		})
	}
}

func (suite *EnvSuite) TestGetWhenPropertiesAreLoadedViaConfigFile() {
	suite.cfgFile = testFile{
		File:    env.DefaultConfigFile(),
		content: suite.flatOuter,
	}
	config.AddPFlags(suite.testFlags)
	suite.verifyEnvCreated()

	suite.verifyEnvContainsValues(suite.flatOuter)
}

func (suite *EnvSuite) TestGetWhenPropertiesAreLoadedViaOSEnvironmentVariables() {
	suite.setEnvVars()
	suite.verifyEnvCreated()

	suite.verifyEnvContainsValues(suite.flatOuter)
}

func (suite *EnvSuite) TestGetOverridePriorityOverPflagSetOverEnvironmentOverFileOverPflagDefault() {
	suite.testFlags.AddFlagSet(singlePFlagSet(key, flagDefaultValue, description))
	suite.verifyEnvCreated()

	suite.Require().Equal(suite.environment.Get(key), flagDefaultValue)

	config.AddPFlags(suite.testFlags)
	suite.cfgFile = testFile{
		File: env.DefaultConfigFile(),
		content: map[string]interface{}{
			key: fileValue,
		},
	}
	suite.verifyEnvCreated()
	suite.Require().Equal(suite.environment.Get(key), fileValue)

	suite.Require().NoError(os.Setenv(strings.ToTitle(key), envValue))
	suite.Require().Equal(suite.environment.Get(key), envValue)

	suite.Require().NoError(suite.testFlags.Set(key, flagValue))
	suite.Require().Equal(suite.environment.Get(key), flagValue)

	suite.environment.Set(key, overrideValue)
	suite.Require().Equal(suite.environment.Get(key), overrideValue)
}

func (suite *EnvSuite) TestSetAddsThePropertyInTheEnvironmentAbstraction() {
	suite.verifyEnvCreated()
	suite.environment.Set(key, overrideValue)

	suite.Require().Equal(suite.environment.Get(key), overrideValue)
}

func (suite *EnvSuite) TestSetHasHighestPriority() {
	suite.testFlags.AddFlagSet(singlePFlagSet(key, flagDefaultValue, description))
	suite.Require().NoError(os.Setenv(key, envValue))
	suite.verifyEnvCreated()
	suite.Require().NoError(suite.testFlags.Set(key, flagValue))

	suite.environment.Set(key, overrideValue)

	suite.Require().Equal(suite.environment.Get(key), overrideValue)
}

func (suite *EnvSuite) TestUnmarshalWhenParameterIsNotAPointerToAStruct() {
	suite.verifyEnvCreated()

	suite.Require().NotNil(suite.environment.Unmarshal(Outer{
		WMapNest: map[string]Nest{
			mapKey: {},
		},
	}))
	suite.Require().NotNil(suite.environment.Unmarshal(10))
}

func (suite *EnvSuite) TestUnmarshalParameterIsAPointerToAStruct() {
	actual := Outer{
		WMapNest: map[string]Nest{
			mapKey: {},
		},
	}

	var tests = []struct {
		Msg       string
		SetUpFunc func()
	}{
		{
			Msg: "when properties are loaded via standard pflags",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(suite.standardPFlagsSet(suite.outer))
			},
		},
		{
			Msg: "when properties are loaded via generated pflags",
			SetUpFunc: func() {
				suite.testFlags.AddFlagSet(generatedPFlagsSet(suite.outer))
			},
		},
		{
			Msg: "when property is loaded via config file",
			SetUpFunc: func() {
				suite.cfgFile = testFile{
					File:    env.DefaultConfigFile(),
					content: suite.flatOuter,
				}
				config.AddPFlags(suite.testFlags)
			},
		},
		{
			Msg: "when properties are loaded via OS environment variables",
			SetUpFunc: func() {
				suite.setEnvVars()
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.Msg, func() {
			defer suite.TearDownTest()
			test.SetUpFunc()

			suite.verifyEnvCreated()

			suite.verifyUnmarshallingIsCorrect(&actual, &suite.outer)
		})
	}
}

func (suite *EnvSuite) TestUnmarshalOverridePriorityOverPflagSetOverEnvironmentOverFileOverPflagDefault() {
	type s struct {
		Key string `mapstructure:"key"`
	}

	str := s{}
	suite.testFlags.AddFlagSet(singlePFlagSet(key, flagDefaultValue, ""))
	suite.verifyEnvCreated()

	suite.verifyUnmarshallingIsCorrect(&str, &s{flagDefaultValue})

	suite.cfgFile = testFile{
		File: env.DefaultConfigFile(),
		content: map[string]interface{}{
			key: fileValue,
		},
	}
	config.AddPFlags(suite.testFlags)
	suite.verifyEnvCreated()
	suite.verifyUnmarshallingIsCorrect(&str, &s{fileValue})

	suite.Require().NoError(os.Setenv(strings.ToTitle(key), envValue))
	suite.verifyUnmarshallingIsCorrect(&str, &s{envValue})
	suite.Require().Equal(suite.environment.Get(key), envValue)

	err := suite.testFlags.Set(key, flagValue)
	suite.Require().NoError(err)
	suite.verifyUnmarshallingIsCorrect(&str, &s{flagValue})

	suite.environment.Set(key, overrideValue)
	suite.verifyUnmarshallingIsCorrect(&str, &s{overrideValue})
}

// Helper functions

func generatedPFlagsSet(s interface{}) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	env.CreatePFlags(set, s)

	return set
}

func (suite *EnvSuite) standardPFlagsSet(o Outer) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)

	suite.walkStructFields(o, func(name string, value interface{}) {
		switch v := value.(type) {
		case bool:
			set.Bool(name, v, description)
		case int:
			set.Int(name, v, description)
		case string:
			set.String(name, v, description)
		case []string:
			set.StringSlice(name, v, description)
		}
	})

	return set
}

func singlePFlagSet(key, defaultValue, description string) *pflag.FlagSet {
	set := pflag.NewFlagSet("testflags", pflag.ExitOnError)
	set.String(key, defaultValue, description)

	return set
}

func (suite *EnvSuite) setPFlags(o Outer) {
	suite.walkStructFields(o, func(name string, value interface{}) {
		switch v := value.(type) {
		case []string:
			suite.Require().NoError(suite.testFlags.Set(name, strings.Join(v, ",")))
		case []interface{}:
			strs := make([]string, 0, 0)
			for _, val := range v {
				strs = append(strs, fmt.Sprint(val))
			}
			suite.Require().NoError(suite.testFlags.Set(name, strings.Join(strs, ",")))
		default:
			suite.Require().NoError(suite.testFlags.Set(name, cast.ToString(v)))
		}
	})
}

func (suite *EnvSuite) setEnvVars() {
	suite.walkStructFields(suite.outer, func(name string, value interface{}) {
		switch v := value.(type) {
		case []string:
			suite.Require().NoError(os.Setenv(strings.ReplaceAll(strings.ToTitle(name), ".", "_"), strings.Join(v, ",")))
		case []interface{}:
			strs := make([]string, 0, 0)
			for _, val := range v {
				strs = append(strs, fmt.Sprint(val))
			}
			suite.Require().NoError(os.Setenv(strings.ReplaceAll(strings.ToTitle(name), ".", "_"), strings.Join(strs, ",")))
		default:
			suite.Require().NoError(os.Setenv(strings.ReplaceAll(strings.ToTitle(name), ".", "_"), cast.ToString(v)))
		}
	})
}

func (suite *EnvSuite) cleanUpEnvVars() {
	suite.walkStructFields(suite.outer, func(name string, value interface{}) {
		suite.Require().NoError(os.Unsetenv(strings.ReplaceAll(strings.ToTitle(name), ".", "_")))
	})
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
	err := os.Remove(f)
	suite.Require().NoError(err)
}

func (suite *EnvSuite) verifyUnmarshallingIsCorrect(actual, expected interface{}) {
	suite.Require().NoError(suite.environment.Unmarshal(actual))
	suite.Require().Equal(actual, expected)
}

func (suite *EnvSuite) verifyEnvCreated() {
	suite.Require().NoError(suite.createEnv())
}

func (suite *EnvSuite) verifyValues(fields map[string]interface{}, prefix string) {
	suite.walk(fields, prefix, func(name string, value interface{}) {
		switch v := value.(type) {
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
	})
}

func (suite *EnvSuite) walkStructFields(s interface{}, operation func(string, interface{})) {
	fields := structs.Map(s)
	suite.walk(fields, "", operation)
}

func (suite *EnvSuite) walk(fields map[string]interface{}, prefix string, operation func(string, interface{})) {
	for name, value := range fields {
		name = strings.ToLower(name)
		switch v := value.(type) {
		case map[string]interface{}:
			suite.walk(v, prefix+name+".", operation)
		default:
			operation(prefix+name, value)
		}
	}
}

func (suite *EnvSuite) verifyEnvContainsValues(expected interface{}) {
	fields := structs.Map(expected)
	suite.verifyValues(fields, "")
}
