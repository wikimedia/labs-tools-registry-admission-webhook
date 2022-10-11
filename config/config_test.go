package config

import (
	"os"
	"reflect"
	"testing"
)

func TestConfigUsesRegistriesDefaultWhenNoEnvVar(t *testing.T) {
	os.Unsetenv("REGISTRIES")

	myConfig, err := GetConfigFromEnv()

	if err != nil {
		t.Errorf("Got error when reading config: %v", err)
	}

	if len(myConfig.Registries) != 1 {
		t.Errorf("Got different than 1 registry %s", &myConfig.Registries)
	}
}

func TestConfigUsesRegistriesFromEnvVar(t *testing.T) {
	defer os.Unsetenv("REGISTRIES")
	os.Setenv("REGISTRIES", "registry1,registry2")

	expectedRegistries := []string{"registry1", "registry2"}

	myConfig, err := GetConfigFromEnv()

	if err != nil {
		t.Errorf("Got error when reading config: %v", err)
	}

	if !reflect.DeepEqual(myConfig.Registries, expectedRegistries) {
		t.Errorf("Got different than 1 registry %s", &myConfig.Registries)
	}
}

func TestConfigReturnsErrorWhenNoRegistriesPassed(t *testing.T) {
	defer os.Unsetenv("REGISTRIES")
	os.Setenv("REGISTRIES", "")

	gottenConfig, err := GetConfigFromEnv()
	if err == nil {
		t.Errorf("Should have gotten error, instead got config: %v", gottenConfig)
	}
}
