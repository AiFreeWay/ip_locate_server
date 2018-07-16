package config

import (
	"../provider"
	"errors"
	"os"
	"strings"
)

const CONFIG_FILE_PATH string = "./config/config.txt"

var ignoredIps []string = nil
var providers []provider.Provider = nil
var db_user = ""
var db_pass = ""

func Init() {
	configFileData := readConfigFile()
	settings := parseConfigData(configFileData)
	installSettings(settings)
}

func GetIgnoredIps() []string {
	if ignoredIps == nil {
		panic(errors.New("Ignored ips not set, call Init() function before or check config.txt"))
	}

	return ignoredIps
}

func GetProviders() []provider.Provider {
	if providers == nil || len(providers) == 0 {
		panic(errors.New("Providers not set, call Init() function before or check config.txt"))
	}

	return providers
}

func GetDbUser() string {
	if db_user == "" {
		panic(errors.New("db_user not set, call Init() function before or check config.txt"))
	}

	return db_user
}

func GetDbPass() string {
	if db_user == "" {
		panic(errors.New("db_pass not set, call Init() function before or check config.txt"))
	}

	return db_pass
}

func readConfigFile() string {
	configFile, openFileErr := os.Open(CONFIG_FILE_PATH)
	if openFileErr != nil {
		panic(openFileErr)
	}

	defer configFile.Close()

	fileStat, getFileStatErr := configFile.Stat()
	if getFileStatErr != nil {
		panic(getFileStatErr)
	}

	buffer := make([]byte, fileStat.Size())
	_, readFileErr := configFile.Read(buffer)
	if readFileErr != nil {
		panic(readFileErr)
	}

	return string(buffer)
}

func parseConfigData(configFileData string) map[string]string {
	mappedSettings := make(map[string]string)

	settings := strings.SplitAfter(configFileData, ";")
	for _, setting := range settings {
		keyValue := strings.SplitAfter(setting, "=")

		if len(keyValue) == 2 {
			mappedSettings[trimConfigParams(keyValue[0])] = trimConfigParams(keyValue[1])
		}
	}

	return mappedSettings
}

func installSettings(settings map[string]string) {
	for settingKey, settingValue := range settings {
		switch settingKey {
		case "ignore_ip":
			ignoreIps := strings.SplitAfter(settingValue, ",")
			if len(ignoreIps) > 0 {
				installIgnoredIps(ignoreIps)
			} else if len(settingValue) > 0 {
				value := [1]string{settingValue}
				installIgnoredIps(value[0:1])
			}
		case "providers":
			providers := strings.SplitAfter(settingValue, ",")
			if len(providers) > 0 {
				installProviders(providers)
			} else if len(settingValue) > 0 {
				value := [1]string{settingValue}
				installProviders(value[0:1])
			}
		case "db_user":
			db_user = settingValue
		case "db_pass":
			db_pass = settingValue
		}
	}
}

func installIgnoredIps(ips []string) {
	for i, ip := range ips {
		ips[i] = strings.Replace(ip, ",", "", -1)
	}

	ignoredIps = ips
}

func installProviders(providersNames []string) {
	for i, provider := range providersNames {
		providersNames[i] = strings.Replace(provider, ",", "", -1)
	}

	providersCount := len(providersNames)
	providersInstances := make([]provider.Provider, providersCount)

	for i, providerName := range providersNames {
		providersInstances[i] = provider.GetProviderByName(providerName)
	}

	providers = providersInstances
}

func trimConfigParams(arg string) string {
	arg = strings.Replace(arg, " ", "", -1)
	arg = strings.Replace(arg, "\n", "", -1)
	arg = strings.Replace(arg, "=", "", -1)
	arg = strings.Replace(arg, ";", "", -1)

	return arg
}
