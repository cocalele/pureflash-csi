/*
Copyright 2023 LiuLele

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"k8s.io/klog/v2"
	"regexp"
)

// PfbdDriver contains the main parameters
type PfbdDriver struct {
	name              string
	nodeID            string
	version           string
	endpoint          string
	hostWritePath     string
	ephemeral         bool
	maxVolumesPerNode int64

	namespace         string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer
}

var (
	vendorVersion = "dev"
)


const (

	fsTypeRegexpString = `TYPE="(\w+)"`
)

var (
	fsTypeRegexp = regexp.MustCompile(fsTypeRegexpString)
)

// NewPfbdDriver creates the driver
func NewPfbdDriver(driverName, nodeID, endpoint string, version string) (*PfbdDriver, error) {
	if driverName == "" {
		return nil, fmt.Errorf("no driver name provided")
	}

	if nodeID == "" {
		return nil, fmt.Errorf("no node id provided")
	}

	if endpoint == "" {
		return nil, fmt.Errorf("no driver endpoint provided")
	}
	if version != "" {
		vendorVersion = version
	}


	klog.Infof("Driver: %v ", driverName)
	klog.Infof("Version: %s", vendorVersion)

	return &PfbdDriver{
		name:              driverName,
		version:           vendorVersion,
		nodeID:            nodeID,
		endpoint:          endpoint,

	}, nil
}

// Run starts the controller plugin
func (driver *PfbdDriver) Run() error {
	var err error
	// Create GRPC servers
	driver.ids = newIdentityServer(driver.name, driver.version)
	driver.ns = newNodeServer(driver.nodeID, driver.ephemeral, driver.maxVolumesPerNode)
	driver.cs, err = newControllerServer(driver.ephemeral, driver.nodeID, driver.hostWritePath, driver.namespace)
	if err != nil {
		return err
	}
	s := newNonBlockingGRPCServer()
	s.start(driver.endpoint, driver.ids, driver.cs, driver.ns)
	s.wait()
	return nil
}




