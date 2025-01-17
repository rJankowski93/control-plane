// Code generated by mockery 2.7.5. DO NOT EDIT.

package mocks

import (
	dberrors "github.com/kyma-project/control-plane/components/provisioner/internal/persistence/dberrors"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-project/control-plane/components/provisioner/internal/model"
)

// ReadSession is an autogenerated mock type for the ReadSession type
type ReadSession struct {
	mock.Mock
}

// GetCluster provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetCluster(runtimeID string) (model.Cluster, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Cluster
	if rf, ok := ret.Get(0).(func(string) model.Cluster); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Cluster)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetGardenerClusterByName provides a mock function with given fields: name
func (_m *ReadSession) GetGardenerClusterByName(name string) (model.Cluster, dberrors.Error) {
	ret := _m.Called(name)

	var r0 model.Cluster
	if rf, ok := ret.Get(0).(func(string) model.Cluster); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(model.Cluster)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(name)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetLastOperation provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetOperation provides a mock function with given fields: operationID
func (_m *ReadSession) GetOperation(operationID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(operationID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(operationID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(operationID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetRuntimeUpgrade provides a mock function with given fields: operationId
func (_m *ReadSession) GetRuntimeUpgrade(operationId string) (model.RuntimeUpgrade, dberrors.Error) {
	ret := _m.Called(operationId)

	var r0 model.RuntimeUpgrade
	if rf, ok := ret.Get(0).(func(string) model.RuntimeUpgrade); ok {
		r0 = rf(operationId)
	} else {
		r0 = ret.Get(0).(model.RuntimeUpgrade)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(operationId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetTenant provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetTenant(runtimeID string) (string, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetTenantForOperation provides a mock function with given fields: operationID
func (_m *ReadSession) GetTenantForOperation(operationID string) (string, dberrors.Error) {
	ret := _m.Called(operationID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(operationID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(operationID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// InProgressOperationsCount provides a mock function with given fields:
func (_m *ReadSession) InProgressOperationsCount() (model.OperationsCount, dberrors.Error) {
	ret := _m.Called()

	var r0 model.OperationsCount
	if rf, ok := ret.Get(0).(func() model.OperationsCount); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(model.OperationsCount)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func() dberrors.Error); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// ListInProgressOperations provides a mock function with given fields:
func (_m *ReadSession) ListInProgressOperations() ([]model.Operation, dberrors.Error) {
	ret := _m.Called()

	var r0 []model.Operation
	if rf, ok := ret.Get(0).(func() []model.Operation); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Operation)
		}
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func() dberrors.Error); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}
