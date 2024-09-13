package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSafeCh(t *testing.T) {
	t.Run("TestInvalidSafeCh", func(t *testing.T) {
		safeCh := &SafeCh[int]{}
		require.NotNil(t, safeCh)

		require.Nil(t, safeCh.ch)

		err := safeCh.Write(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "channel not initialized. use MakeSafeCh")

		err = safeCh.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "channel not initialized. use MakeSafeCh")

		ch := safeCh.ReadCh()
		require.NotNil(t, ch)
	})

	t.Run("TestMakeSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)
	})

	t.Run("TestCloseSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Close()
		require.NoError(t, err)
	})

	t.Run("TestCloseCloseSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Close()
		require.NoError(t, err)

		err = safeCh.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "channel already closed")
	})

	t.Run("TestWriteSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Write(1)
		require.NoError(t, err)
	})

	t.Run("TestWriteCloseSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Write(1)
		require.NoError(t, err)

		err = safeCh.Close()
		require.NoError(t, err)
	})

	t.Run("TestWriteCloseWriteSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Write(1)
		require.NoError(t, err)

		err = safeCh.Close()
		require.NoError(t, err)

		err = safeCh.Write(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "trying to write to a closed channel")
	})

	t.Run("TestReadSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		go func(t *testing.T, sch *SafeCh[int]) {
			t.Helper()

			err := sch.Write(1)
			require.NoError(t, err)
		}(t, safeCh)

		value, ok := <-safeCh.ReadCh()
		require.True(t, ok)
		require.Equal(t, 1, value)
	})

	t.Run("TestReadCloseSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		go func(t *testing.T, sch *SafeCh[int]) {
			t.Helper()

			err := sch.Write(1)
			require.NoError(t, err)
		}(t, safeCh)

		value, ok := <-safeCh.ReadCh()
		require.True(t, ok)
		require.Equal(t, 1, value)

		err := safeCh.Close()
		require.NoError(t, err)
	})

	t.Run("TestCloseReadSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Close()
		require.NoError(t, err)

		_, ok := <-safeCh.ReadCh()
		require.False(t, ok)
	})

	t.Run("TestComplexSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		go func(t *testing.T, sch *SafeCh[int]) {
			t.Helper()

			<-time.After(time.Millisecond * 100)

			err := sch.Write(1)
			require.NoError(t, err)

			<-time.After(time.Millisecond * 100)

			err = sch.Close()
			require.NoError(t, err)
		}(t, safeCh)

		firstIteration := true

		for {
			select {
			case value, ok := <-safeCh.ReadCh():
				if firstIteration {
					require.True(t, ok)
					require.Equal(t, 1, value)

					firstIteration = false
				} else {
					require.False(t, ok)

					return
				}
			case <-time.After(time.Millisecond * 300):
				t.Fatalf("timeout")

				return
			}
		}
	})
}
