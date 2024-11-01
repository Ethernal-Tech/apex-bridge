package common

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeRegister(t *testing.T) {
	type dummyStruct struct {
		_ float64
	}

	int64Type := reflect.TypeOf(int64(0))
	dummyType := reflect.TypeOf(dummyStruct{})
	selfType := reflect.TypeOf(typeRegisterImpl{})

	tr := NewTypeRegister()

	tr.SetType("a", int64Type)
	tr.SetType("b", dummyType)
	tr.SetType("c", selfType)

	v, err := tr.GetType("a")
	require.NoError(t, err)
	require.Equal(t, int64Type, v)

	v, err = tr.GetType("b")
	require.NoError(t, err)
	require.Equal(t, dummyType, v)

	v, err = tr.GetType("c")
	require.NoError(t, err)
	require.Equal(t, selfType, v)

	_, err = tr.GetType("d")
	require.Error(t, err)

	val1, err := GetRegisteredTypeInstance[int64](tr, "a")
	require.NoError(t, err)
	require.Equal(t, int64(0), val1)

	val2, err := GetRegisteredTypeInstance[dummyStruct](tr, "b")
	require.NoError(t, err)
	require.Equal(t, dummyStruct{}, val2)

	val3, err := GetRegisteredTypeInstance[*dummyStruct](tr, "b")
	require.NoError(t, err)
	require.Equal(t, &dummyStruct{}, val3)

	val4, err := GetRegisteredTypeInstance[TypeRegister](tr, "c")
	require.NoError(t, err)
	require.Equal(t, &typeRegisterImpl{}, val4)

	_, err = GetRegisteredTypeInstance[float64](tr, "a")
	require.Error(t, err)
}
