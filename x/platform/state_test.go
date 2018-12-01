package platform

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClient struct {
	Num int
	Buf []byte
}

func (tcl testClient) Update(_ *Context) error {
	return nil
}

func TestPlatform_state_reload(t *testing.T) {
	var save []byte
	require.NoError(t, func() error {
		p, err := New(nil, nil, Config{})
		require.NoError(t, err)
		var cl testClient
		p.TimingEnabled = true
		p.client = &cl
		cl.Num = 42
		cl.Buf = append(cl.Buf, "hello"...)
		var buf bytes.Buffer
		if err := p.writeState(&buf); err != nil {
			return err
		}
		save = buf.Bytes()

		return nil
	}(), "unexpected writeState error")
	require.NoError(t, func() error {
		p, err := New(nil, nil, Config{})
		require.NoError(t, err)
		var cl testClient
		p.client = &cl
		if err := p.readState(bytes.NewReader(save)); err != nil {
			return err
		}
		assert.True(t, p.TimingEnabled)
		assert.Equal(t, 42, cl.Num)
		assert.Equal(t, "hello", string(cl.Buf))
		return nil
	}(), "unexpected readState error")
}

func TestPlatform_state_rewind(t *testing.T) {
	require.NoError(t, func() error {
		p, err := New(nil, nil, Config{})
		require.NoError(t, err)
		var cl testClient
		p.TimingEnabled = true
		p.client = &cl
		cl.Num = 42
		cl.Buf = append(cl.Buf, "hello"...)

		var buf bytes.Buffer
		if err := p.writeState(&buf); err != nil {
			return fmt.Errorf("failed first save: %v", err)
		}
		save1 := append([]byte(nil), buf.Bytes()...)

		cl.Num++
		cl.Buf = append(cl.Buf, " world"...)
		p.TimingEnabled = false

		buf.Reset()
		if err := p.writeState(&buf); err != nil {
			return fmt.Errorf("failed second save: %v", err)
		}
		save2 := append([]byte(nil), buf.Bytes()...)

		// load save1
		if err := p.readState(bytes.NewReader(save1)); err != nil {
			return fmt.Errorf("failed first read: %v", err)
		}
		assert.False(t, p.LogTiming)
		assert.True(t, p.TimingEnabled)
		assert.Equal(t, 42, cl.Num)
		assert.Equal(t, "hello", string(cl.Buf))

		// load save2
		if err := p.readState(bytes.NewReader(save2)); err != nil {
			return fmt.Errorf("failed second read: %v", err)
		}
		assert.False(t, p.LogTiming)
		assert.False(t, p.TimingEnabled)
		assert.Equal(t, 43, cl.Num)
		assert.Equal(t, "hello world", string(cl.Buf))

		// load save1
		if err := p.readState(bytes.NewReader(save1)); err != nil {
			return fmt.Errorf("failed third read: %v", err)
		}
		assert.False(t, p.LogTiming)
		assert.True(t, p.TimingEnabled)
		assert.Equal(t, 42, cl.Num)
		assert.Equal(t, "hello", string(cl.Buf))

		return nil
	}(), "unexpected write/readState error")

}
