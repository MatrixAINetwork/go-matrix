// Copyright 2018 The MATRIX Authors as well as Copyright 2014-2017 The go-ethereum Authors
// This file is consisted of the MATRIX library and part of the go-ethereum library.
//
// The MATRIX-ethereum library is free software: you can redistribute it and/or modify it under the terms of the MIT License.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, 
//and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject tothe following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, 
//WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISINGFROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
//OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
package goupnp

import (
	"fmt"
	"net/url"

	"github.com/huin/goupnp/soap"
)

// ServiceClient is a SOAP client, root device and the service for the SOAP
// client rolled into one value. The root device, location, and service are
// intended to be informational. Location can be used to later recreate a
// ServiceClient with NewServiceClientByURL if the service is still present;
// bypassing the discovery process.
type ServiceClient struct {
	SOAPClient *soap.SOAPClient
	RootDevice *RootDevice
	Location   *url.URL
	Service    *Service
}

// NewServiceClients discovers services, and returns clients for them. err will
// report any error with the discovery process (blocking any device/service
// discovery), errors reports errors on a per-root-device basis.
func NewServiceClients(searchTarget string) (clients []ServiceClient, errors []error, err error) {
	var maybeRootDevices []MaybeRootDevice
	if maybeRootDevices, err = DiscoverDevices(searchTarget); err != nil {
		return
	}

	clients = make([]ServiceClient, 0, len(maybeRootDevices))

	for _, maybeRootDevice := range maybeRootDevices {
		if maybeRootDevice.Err != nil {
			errors = append(errors, maybeRootDevice.Err)
			continue
		}

		deviceClients, err := NewServiceClientsFromRootDevice(maybeRootDevice.Root, maybeRootDevice.Location, searchTarget)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		clients = append(clients, deviceClients...)
	}

	return
}

// NewServiceClientsByURL creates client(s) for the given service URN, for a
// root device at the given URL.
func NewServiceClientsByURL(loc *url.URL, searchTarget string) ([]ServiceClient, error) {
	rootDevice, err := DeviceByURL(loc)
	if err != nil {
		return nil, err
	}
	return NewServiceClientsFromRootDevice(rootDevice, loc, searchTarget)
}

// NewServiceClientsFromDevice creates client(s) for the given service URN, in
// a given root device. The loc parameter is simply assigned to the
// Location attribute of the returned ServiceClient(s).
func NewServiceClientsFromRootDevice(rootDevice *RootDevice, loc *url.URL, searchTarget string) ([]ServiceClient, error) {
	device := &rootDevice.Device
	srvs := device.FindService(searchTarget)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("goupnp: service %q not found within device %q (UDN=%q)",
			searchTarget, device.FriendlyName, device.UDN)
	}

	clients := make([]ServiceClient, 0, len(srvs))
	for _, srv := range srvs {
		clients = append(clients, ServiceClient{
			SOAPClient: srv.NewSOAPClient(),
			RootDevice: rootDevice,
			Location:   loc,
			Service:    srv,
		})
	}
	return clients, nil
}

// GetServiceClient returns the ServiceClient itself. This is provided so that the
// service client attributes can be accessed via an interface method on a
// wrapping type.
func (client *ServiceClient) GetServiceClient() *ServiceClient {
	return client
}
