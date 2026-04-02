package rfplayer

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

type mockPort struct {
	writeBuf []byte
	readResp string
	readErr  error
	writeErr error
	closed   bool
}

func (m *mockPort) Read(p []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	n = copy(p, m.readResp)
	return n, nil
}

func (m *mockPort) Write(p []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.writeBuf = append(m.writeBuf, p...)
	return len(p), nil
}

func (m *mockPort) Close() error {
	m.closed = true
	return nil
}

func (m *mockPort) Flush() error {
	return nil
}

func (m *mockPort) written() string {
	return string(m.writeBuf)
}

func TestSendCommand(t *testing.T) {
	tests := []struct {
		name       string
		cmd        string
		resp       string
		readErr    error
		writeErr   error
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "simple command",
			cmd:        "PING",
			resp:       "ZIA--PONG\r",
			wantOutput: "PONG",
		},
		{
			name:       "response without prefix",
			cmd:        "HELLO",
			resp:       "RFPlayer v1.15\r",
			wantOutput: "RFPlayer v1.15",
		},
		{
			name:       "response with whitespace",
			cmd:        "STATUS SYSTEM TEXT",
			resp:       "ZIA--OK\r\n",
			wantOutput: "OK",
		},
		{
			name:    "read error",
			cmd:     "PING",
			readErr: errors.New("read timeout"),
			wantErr: true,
		},
		{
			name:     "write error",
			cmd:      "PING",
			writeErr: errors.New("write failed"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockPort{
				readResp: tt.resp,
				readErr:  tt.readErr,
				writeErr: tt.writeErr,
			}

			rf := NewWithPort(mp)
			output, err := rf.SendCommand(tt.cmd)

			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			assert.Equal(t, tt.wantOutput, output)

			expectedWrite := fmt.Sprintf("ZIA++%s\r", tt.cmd)
			assert.Assert(t, strings.Contains(mp.written(), expectedWrite),
				"expected write to contain %q, got %q", expectedWrite, mp.written())
		})
	}
}

func TestPing(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--PONG\r"}
		rf := NewWithPort(mp)
		assert.NilError(t, rf.Ping())
	})

	t.Run("unexpected response", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--ERROR\r"}
		rf := NewWithPort(mp)
		assert.ErrorContains(t, rf.Ping(), "unexpected response")
	})

	t.Run("read error", func(t *testing.T) {
		mp := &mockPort{readErr: errors.New("timeout")}
		rf := NewWithPort(mp)
		assert.ErrorContains(t, rf.Ping(), "failed to send PING command")
	})
}

func TestHello(t *testing.T) {
	t.Run("successful hello", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--RFPlayer v1.15\r"}
		rf := NewWithPort(mp)
		resp, err := rf.Hello()
		assert.NilError(t, err)
		assert.Equal(t, "RFPlayer v1.15", resp)
	})

	t.Run("read error", func(t *testing.T) {
		mp := &mockPort{readErr: errors.New("timeout")}
		rf := NewWithPort(mp)
		_, err := rf.Hello()
		assert.ErrorContains(t, err, "failed to send HELLO command")
	})
}

func TestEmitSignal(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.EmitSignal("CH5", 42, "ON")
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++ON ID 42 CH5\r"))
}

func TestRecordSignal(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.RecordSignal(1, "ON", "zonwering-1-1")
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++PARROTLEARN ID 1 ON [zonwering-1-1]\r"))
}

func TestSetFrequency(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.SetFrequency("H", 868350)
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++FREQ H 868350\r"))
}

func TestEnableReceiver(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.EnableReceiver("CH5", "CH4")
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++RECEIVER + CH5 CH4\r"))
}

func TestSetFormat(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.SetFormat("JSON")
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++FORMAT JSON\r"))
}

func TestGetStatus(t *testing.T) {
	t.Run("default system text", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--OK\r"}
		rf := NewWithPort(mp)

		resp, err := rf.GetStatus("", "")
		assert.NilError(t, err)
		assert.Equal(t, "OK", resp)
		assert.Assert(t, strings.Contains(mp.written(), "ZIA++STATUS SYSTEM TEXT\r"))
	})

	t.Run("custom type and format", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--data\r"}
		rf := NewWithPort(mp)

		resp, err := rf.GetStatus("PARROT", "JSON")
		assert.NilError(t, err)
		assert.Equal(t, "data", resp)
		assert.Assert(t, strings.Contains(mp.written(), "ZIA++STATUS PARROT JSON\r"))
	})
}

func TestParrotRemapping(t *testing.T) {
	mp := &mockPort{readResp: "ZIA--OK\r"}
	rf := NewWithPort(mp)

	resp, err := rf.ParrotRemapping("CH5", 1)
	assert.NilError(t, err)
	assert.Equal(t, "OK", resp)
	assert.Assert(t, strings.Contains(mp.written(), "ZIA++REMAPPING PARROT ONOFF CH5 A1\r"))
}

func TestFactoryReset(t *testing.T) {
	t.Run("partial reset", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--OK\r"}
		rf := NewWithPort(mp)
		assert.NilError(t, rf.FactoryReset(false))
		assert.Assert(t, strings.Contains(mp.written(), "ZIA++FACTORYRESET\r"))
	})

	t.Run("full reset", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--OK\r"}
		rf := NewWithPort(mp)
		assert.NilError(t, rf.FactoryReset(true))
		assert.Assert(t, strings.Contains(mp.written(), "ZIA++FACTORYRESET ALL\r"))
	})

	t.Run("unexpected response", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--ERROR\r"}
		rf := NewWithPort(mp)
		assert.ErrorContains(t, rf.FactoryReset(false), "unexpected response")
	})
}

func TestIdToX10(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{1, "A1"},
		{16, "A16"},
		{17, "B1"},
		{32, "B16"},
		{48, "C16"},
		{256, "P16"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, idToX10(tt.id))
		})
	}
}

func TestClose(t *testing.T) {
	mp := &mockPort{}
	rf := NewWithPort(mp)
	assert.NilError(t, rf.Close())
	assert.Assert(t, mp.closed)
}
