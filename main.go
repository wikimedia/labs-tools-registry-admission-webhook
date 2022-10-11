package main

import (
	"gerrit.wikimedia.org/labs/tools/registry-admission-webhook/config"
	"gerrit.wikimedia.org/labs/tools/registry-admission-webhook/server"
	"github.com/sirupsen/logrus"
)

// Config is the general configuration of the webhook via env variables
func main() {
	myConfig, err := config.GetConfigFromEnv()
	if err != nil {
		logrus.Error("Got malformed configuration from environment: ", err)
	}

	if myConfig.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infoln(myConfig)
	nsac := server.RegistryAdmission{Registries: myConfig.Registries}
	s := server.GetAdmissionValidationServer(&nsac, myConfig.TLSCert, myConfig.TLSKey, myConfig.ListenOn)
	s.ListenAndServeTLS("", "")
}
