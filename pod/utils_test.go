// Copyright (c) 2018 The MATRIX Authors
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php

// Contains a batch of utility type declarations used by the tests. As the node
// operates on unique types, a lot of them are needed to check various features.

package pod

import (
	"reflect"

	"github.com/MatrixAINetwork/go-matrix/p2p"
	"github.com/MatrixAINetwork/go-matrix/rpc"
)

// NoopService is a trivial implementation of the Service interface.
type NoopService struct{}

func (s *NoopService) Protocols() []p2p.Protocol { return nil }
func (s *NoopService) APIs() []rpc.API           { return nil }
func (s *NoopService) Start(*p2p.Server) error   { return nil }
func (s *NoopService) Stop() error               { return nil }

func NewNoopService(*ServiceContext) (Service, error) { return new(NoopService), nil }

// Set of services all wrapping the base NoopService resulting in the same method
// signatures but different outer types.
type NoopServiceA struct{ NoopService }
type NoopServiceB struct{ NoopService }
type NoopServiceC struct{ NoopService }

func NewNoopServiceA(*ServiceContext) (Service, error) { return new(NoopServiceA), nil }
func NewNoopServiceB(*ServiceContext) (Service, error) { return new(NoopServiceB), nil }
func NewNoopServiceC(*ServiceContext) (Service, error) { return new(NoopServiceC), nil }

// InstrumentedService is an implementation of Service for which all interface
// methods can be instrumented both return value as well as event hook wise.
type InstrumentedService struct {
	protocols []p2p.Protocol
	apis      []rpc.API
	start     error
	stop      error

	protocolsHook func()
	startHook     func(*p2p.Server)
	stopHook      func()
}

func NewInstrumentedService(*ServiceContext) (Service, error) { return new(InstrumentedService), nil }

func (s *InstrumentedService) Protocols() []p2p.Protocol {
	if s.protocolsHook != nil {
		s.protocolsHook()
	}
	return s.protocols
}

func (s *InstrumentedService) APIs() []rpc.API {
	return s.apis
}

func (s *InstrumentedService) Start(server *p2p.Server) error {
	if s.startHook != nil {
		s.startHook(server)
	}
	return s.start
}

func (s *InstrumentedService) Stop() error {
	if s.stopHook != nil {
		s.stopHook()
	}
	return s.stop
}

// InstrumentingWrapper is a method to specialize a service constructor returning
// a generic InstrumentedService into one returning a wrapping specific one.
type InstrumentingWrapper func(base ServiceConstructor) ServiceConstructor

func InstrumentingWrapperMaker(base ServiceConstructor, kind reflect.Type) ServiceConstructor {
	return func(ctx *ServiceContext) (Service, error) {
		obj, err := base(ctx)
		if err != nil {
			return nil, err
		}
		wrapper := reflect.New(kind)
		wrapper.Elem().Field(0).Set(reflect.ValueOf(obj).Elem())

		return wrapper.Interface().(Service), nil
	}
}

// Set of services all wrapping the base InstrumentedService resulting in the
// same method signatures but different outer types.
type InstrumentedServiceA struct{ InstrumentedService }
type InstrumentedServiceB struct{ InstrumentedService }
type InstrumentedServiceC struct{ InstrumentedService }

func InstrumentedServiceMakerA(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceA{}))
}

func InstrumentedServiceMakerB(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceB{}))
}

func InstrumentedServiceMakerC(base ServiceConstructor) ServiceConstructor {
	return InstrumentingWrapperMaker(base, reflect.TypeOf(InstrumentedServiceC{}))
}

// OneMethodAPI is a single-method API handler to be returned by test services.
type OneMethodAPI struct {
	fun func()
}

func (api *OneMethodAPI) TheOneMethod() {
	if api.fun != nil {
		api.fun()
	}
}
