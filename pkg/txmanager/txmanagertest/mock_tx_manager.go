
package txmanagertest

import (
	context "context"

	gorm "gorm.io/gorm"

	mock "github.com/stretchr/testify/mock"

	txcore "github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

type MockTxManager struct {
	mock.Mock
}

type MockTxManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTxManager) EXPECT() *MockTxManager_Expecter {
	return &MockTxManager_Expecter{mock: &_m.Mock}
}

func (_m *MockTxManager) Do(ctx context.Context, txName string, fn func(context.Context) error, opts ...txcore.Option) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, txName, fn)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Do")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, func(context.Context) error, ...txcore.Option) error); ok {
		r0 = rf(ctx, txName, fn, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type MockTxManager_Do_Call struct {
	*mock.Call
}

func (_e *MockTxManager_Expecter) Do(ctx interface{}, txName interface{}, fn interface{}, opts ...interface{}) *MockTxManager_Do_Call {
	return &MockTxManager_Do_Call{Call: _e.mock.On("Do",
		append([]interface{}{ctx, txName, fn}, opts...)...)}
}

func (_c *MockTxManager_Do_Call) Run(run func(ctx context.Context, txName string, fn func(context.Context) error, opts ...txcore.Option)) *MockTxManager_Do_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]txcore.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(txcore.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(func(context.Context) error), variadicArgs...)
	})
	return _c
}

func (_c *MockTxManager_Do_Call) Return(_a0 error) *MockTxManager_Do_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTxManager_Do_Call) RunAndReturn(run func(context.Context, string, func(context.Context) error, ...txcore.Option) error) *MockTxManager_Do_Call {
	_c.Call.Return(run)
	return _c
}

func (_m *MockTxManager) GetConnection(ctx context.Context) *gorm.DB {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetConnection")
	}

	var r0 *gorm.DB
	if rf, ok := ret.Get(0).(func(context.Context) *gorm.DB); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	return r0
}

type MockTxManager_GetConnection_Call struct {
	*mock.Call
}

func (_e *MockTxManager_Expecter) GetConnection(ctx interface{}) *MockTxManager_GetConnection_Call {
	return &MockTxManager_GetConnection_Call{Call: _e.mock.On("GetConnection", ctx)}
}

func (_c *MockTxManager_GetConnection_Call) Run(run func(ctx context.Context)) *MockTxManager_GetConnection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTxManager_GetConnection_Call) Return(_a0 *gorm.DB) *MockTxManager_GetConnection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTxManager_GetConnection_Call) RunAndReturn(run func(context.Context) *gorm.DB) *MockTxManager_GetConnection_Call {
	_c.Call.Return(run)
	return _c
}

func (_m *MockTxManager) GetRawConnection(ctx context.Context) *gorm.DB {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetRawConnection")
	}

	var r0 *gorm.DB
	if rf, ok := ret.Get(0).(func(context.Context) *gorm.DB); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	return r0
}

type MockTxManager_GetRawConnection_Call struct {
	*mock.Call
}

func (_e *MockTxManager_Expecter) GetRawConnection(ctx interface{}) *MockTxManager_GetRawConnection_Call {
	return &MockTxManager_GetRawConnection_Call{Call: _e.mock.On("GetRawConnection", ctx)}
}

func (_c *MockTxManager_GetRawConnection_Call) Run(run func(ctx context.Context)) *MockTxManager_GetRawConnection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTxManager_GetRawConnection_Call) Return(_a0 *gorm.DB) *MockTxManager_GetRawConnection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTxManager_GetRawConnection_Call) RunAndReturn(run func(context.Context) *gorm.DB) *MockTxManager_GetRawConnection_Call {
	_c.Call.Return(run)
	return _c
}

func (_m *MockTxManager) InTransaction(ctx context.Context) bool {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for InTransaction")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type MockTxManager_InTransaction_Call struct {
	*mock.Call
}

func (_e *MockTxManager_Expecter) InTransaction(ctx interface{}) *MockTxManager_InTransaction_Call {
	return &MockTxManager_InTransaction_Call{Call: _e.mock.On("InTransaction", ctx)}
}

func (_c *MockTxManager_InTransaction_Call) Run(run func(ctx context.Context)) *MockTxManager_InTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockTxManager_InTransaction_Call) Return(_a0 bool) *MockTxManager_InTransaction_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTxManager_InTransaction_Call) RunAndReturn(run func(context.Context) bool) *MockTxManager_InTransaction_Call {
	_c.Call.Return(run)
	return _c
}

func NewMockTxManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTxManager {
	mock := &MockTxManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
