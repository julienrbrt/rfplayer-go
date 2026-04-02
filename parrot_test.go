package rfplayer

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetParrotDevices(t *testing.T) {
	t.Run("successful parse", func(t *testing.T) {
		resp := `{"parrotStatus":{"entry1":{"reminder":"zonwering-1-1","protocol":"CH5"},"entry2":{"reminder":"zonwering-1-2","protocol":"CH5"}}}`
		mp := &mockPort{readResp: "ZIA--" + resp + "\r"}
		rf := NewWithPort(mp)

		devices, err := GetParrotDevices(rf)
		assert.NilError(t, err)

		assert.Equal(t, 2, len(devices))

		byID := make(map[int]ParrotDevice)
		for _, d := range devices {
			byID[d.ID] = d
		}
		assert.Equal(t, "zonwering-1-1", byID[1].Name)
		assert.Equal(t, "CH5", byID[1].Protocol)
		assert.Equal(t, "zonwering-1-2", byID[2].Name)
	})

	t.Run("empty parrot status", func(t *testing.T) {
		resp := `{"parrotStatus":{}}`
		mp := &mockPort{readResp: "ZIA--" + resp + "\r"}
		rf := NewWithPort(mp)

		devices, err := GetParrotDevices(rf)
		assert.NilError(t, err)
		assert.Equal(t, 0, len(devices))
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--not-json\r"}
		rf := NewWithPort(mp)

		_, err := GetParrotDevices(rf)
		assert.ErrorContains(t, err, "failed to parse Parrot status")
	})

	t.Run("missing parrotStatus key", func(t *testing.T) {
		mp := &mockPort{readResp: "ZIA--{\"other\":{}}\r"}
		rf := NewWithPort(mp)

		_, err := GetParrotDevices(rf)
		assert.ErrorContains(t, err, "invalid Parrot status format")
	})

	t.Run("non-entry keys are ignored", func(t *testing.T) {
		resp := `{"parrotStatus":{"version":"1.0","entry3":{"reminder":"lamp","protocol":"CH4"}}}`
		mp := &mockPort{readResp: "ZIA--" + resp + "\r"}
		rf := NewWithPort(mp)

		devices, err := GetParrotDevices(rf)
		assert.NilError(t, err)
		assert.Equal(t, 1, len(devices))
		assert.Equal(t, 3, devices[0].ID)
		assert.Equal(t, "lamp", devices[0].Name)
	})

	t.Run("entry with invalid value is skipped", func(t *testing.T) {
		resp := `{"parrotStatus":{"entry1":"not-a-map","entry2":{"reminder":"ok","protocol":"CH5"}}}`
		mp := &mockPort{readResp: "ZIA--" + resp + "\r"}
		rf := NewWithPort(mp)

		devices, err := GetParrotDevices(rf)
		assert.NilError(t, err)
		assert.Equal(t, 1, len(devices))
		assert.Equal(t, "ok", devices[0].Name)
	})
}
