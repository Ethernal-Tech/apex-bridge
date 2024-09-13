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

		ch, err := safeCh.ReadCh()
		require.Nil(t, ch)
		require.Error(t, err)
		require.ErrorContains(t, err, "channel not initialized. use MakeSafeCh")

		err = safeCh.Write(1)
		require.Error(t, err)
		require.ErrorContains(t, err, "channel not initialized. use MakeSafeCh")

		err = safeCh.Close()
		require.Error(t, err)
		require.ErrorContains(t, err, "channel not initialized. use MakeSafeCh")
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

		ch, err := safeCh.ReadCh()
		require.NotNil(t, ch)
		require.NoError(t, err)

		value, ok := <-ch
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

		ch, err := safeCh.ReadCh()
		require.NotNil(t, ch)
		require.NoError(t, err)

		value, ok := <-ch
		require.True(t, ok)
		require.Equal(t, 1, value)

		err = safeCh.Close()
		require.NoError(t, err)
	})

	t.Run("TestCloseReadSafeCh", func(t *testing.T) {
		safeCh := MakeSafeCh[int](1)
		require.NotNil(t, safeCh)

		err := safeCh.Close()
		require.NoError(t, err)

		ch, err := safeCh.ReadCh()
		require.NotNil(t, ch)
		require.NoError(t, err)

		_, ok := <-ch
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

		ch, err := safeCh.ReadCh()
		require.NotNil(t, ch)
		require.NoError(t, err)

		for {
			select {
			case value, ok := <-ch:
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
