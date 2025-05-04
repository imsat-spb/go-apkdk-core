package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSpecialDeviceForHost(t *testing.T) {
	specId := GetSpecialDeviceForHost(500)
	assert.Equal(t, int32(500+MaxDeviceId), specId)
}

func TestGetHostForSpecialDevice(t *testing.T) {
	tests := []struct {
		name    string
		hostId  int
		specId  int32
		isValid bool
	}{
		{name: "NotValidSpecialDevice", isValid: false, specId: 1000},
		{name: "ValidSpecialDevice", isValid: true, hostId: 1000, specId: 1000 + MaxDeviceId},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hostId, err := GetHostForSpecialDevice(test.specId)

			if !test.isValid {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.hostId, hostId)
			}
		})
	}
}

func TestNotSupportedBitsConversionFunction(t *testing.T) {
	f, err := GetDataConverterFunction(8)
	assert.Nil(t, f)
	assert.NotNil(t, err)
}

func TestConversionFrom32BitFunction(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		sensorId uint16
		expected float32
	}{
		{
			name:     "test1",
			sensorId: 0,
			data:     []byte{0, 0, 0, 0},
			expected: 0.0,
		},

		{
			name:     "test1_2",
			sensorId: 1,
			data:     []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: 0.0,
		},
		{
			name:     "test2",
			sensorId: 0,
			data:     []byte{1, 0, 0, 0},
			expected: 0.001,
		},

		{
			name:     "test2_2",
			sensorId: 1,
			data:     []byte{0, 0, 0, 0, 1, 0, 0, 0},
			expected: 0.001,
		},

		{
			name:     "test3",
			sensorId: 0,
			data:     []byte{255, 0, 0, 0},
			expected: 0.255,
		},

		{
			name:     "test3_2",
			sensorId: 1,
			data:     []byte{0, 0, 0, 0, 255, 0, 0, 0},
			expected: 0.255,
		},

		{
			name:     "test4",
			sensorId: 0,
			data:     []byte{0x00, 0x00, 0x00, 0x80}, // undefined
			expected: float32NaN,
		},

		{
			name:     "test5",
			sensorId: 0,
			data:     []byte{}, // out of range
			expected: float32NaN,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := GetDataConverterFunction(32)
			assert.Nil(t, err)

			val := f(test.data, test.sensorId)

			if IsNaN(val) {
				assert.True(t, IsNaN(test.expected))
			} else {
				assert.Equal(t, test.expected, val)
			}

		})
	}
}

func TestConversionFrom16BitFunction(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		sensorId uint16
		expected float32
	}{
		{
			name:     "test1",
			sensorId: 0,
			data:     []byte{0, 0},
			expected: 0.0,
		},

		{
			name:     "test1_2",
			sensorId: 1,
			data:     []byte{0, 0, 0, 0},
			expected: 0.0,
		},
		{
			name:     "test2",
			sensorId: 0,
			data:     []byte{1, 0},
			expected: 0.001,
		},

		{
			name:     "test2_2",
			sensorId: 1,
			data:     []byte{0, 0, 1, 0},
			expected: 0.001,
		},

		{
			name:     "test3",
			sensorId: 0,
			data:     []byte{255, 0},
			expected: 0.255,
		},

		{
			name:     "test3_2",
			sensorId: 1,
			data:     []byte{0, 0, 255, 0},
			expected: 0.255,
		},

		{
			name:     "test4",
			sensorId: 0,
			data:     []byte{0x00, 0x80}, // undefined
			expected: float32NaN,
		},

		{
			name:     "test5",
			sensorId: 0,
			data:     []byte{}, // out of range
			expected: float32NaN,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := GetDataConverterFunction(16)
			assert.Nil(t, err)

			val := f(test.data, test.sensorId)

			if IsNaN(val) {
				assert.True(t, IsNaN(test.expected))
			} else {
				assert.Equal(t, test.expected, val)
			}

		})
	}
}
