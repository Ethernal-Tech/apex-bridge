package common

import (
	"fmt"
	"reflect"
	"sync"
)

type TypeRegister interface {
	GetType(key string) (reflect.Type, error)
	SetType(key string, registedType reflect.Type)
}

type typeRegisterImpl struct {
	lock sync.Mutex
	data map[string]reflect.Type
}

var _ TypeRegister = (*typeRegisterImpl)(nil)

func NewTypeRegister() TypeRegister {
	return &typeRegisterImpl{
		data: map[string]reflect.Type{},
	}
}

func (r *typeRegisterImpl) SetType(key string, registedType reflect.Type) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.data[key] = registedType
}

func (r *typeRegisterImpl) GetType(key string) (reflect.Type, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	t, exists := r.data[key]
	if !exists {
		return nil, fmt.Errorf("type not registered: %s", key)
	}

	return t, nil
}

func GetRegisteredTypeInstance[T any](register TypeRegister, key string) (result T, err error) {
	t, err := register.GetType(key)
	if err != nil {
		return result, err
	}

	var (
		ok   bool
		newT = reflect.New(t)
	)

	result, ok = newT.Interface().(T)
	if !ok {
		result, ok = newT.Elem().Interface().(T)
		if !ok {
			return result, fmt.Errorf("failed to convert type: %s", key)
		}
	}

	return result, nil
}
