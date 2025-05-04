package core

import (
	"encoding/binary"
	"fmt"
	"math"
)

var (
	float32NaN = math.Float32frombits(UndefinedMeasureValue)
)

func GetNaN() float32 {
	return float32NaN
}

func IsNaN(x float32) bool {
	return math.IsNaN(float64(x))
}

func MakeFloat32FromUint32(x uint32) float32 {
	if x == SystemUndefined32BitValue {
		return float32NaN
	}
	// Проверяем знак и если это отрицательное число, то переводим из дополнительного кода
	isNegative := false
	if x&0x80000000 != 0 {
		x = ^x + 1
		isNegative = true
	}

	aVal := float32(x/1000) + float32(x%1000)*0.001

	if isNegative {
		aVal = -aVal
	}
	return round3(aVal)
}

func makeFloat32FromUint16(x uint16) float32 {
	// Полученное значение необходимо разделить на 1000, если это не неопределенное значение
	if x == SystemUndefined16BitValue {
		return float32NaN
	}

	// Проверяем знак и если это отрицательное число, то переводим из дополнительного кода
	isNegative := false
	if x&0x8000 != 0 {
		x = ^x + 1
		isNegative = true
	}

	aVal := float32(x/1000) + float32(x%1000)*0.001

	if isNegative {
		aVal = -aVal
	}
	return round3(aVal)
}

func round3(x float32) float32 {
	return float32(math.Round(float64(x*1000)) / 1000)
}

const MaxDeviceId = 0x20000000

type GetDataFunction func(data []byte, sensorId uint16) float32

func GetDataConverterFunction(bitsPerSensor byte) (GetDataFunction, error) {

	var handler GetDataFunction

	switch bitsPerSensor {
	case 16:
		handler = func(data []byte, sensorId uint16) float32 {
			if int(sensorId) >= len(data)/2 {
				return float32NaN
			}
			uint16Value := binary.LittleEndian.Uint16(data[sensorId*2:])
			return makeFloat32FromUint16(uint16Value)
		}
	case 32:
		handler = func(data []byte, sensorId uint16) float32 {
			if int(sensorId) >= len(data)/4 {
				return float32NaN
			}
			uint32Value := binary.LittleEndian.Uint32(data[sensorId*4:])
			return MakeFloat32FromUint32(uint32Value)
		}
	default:
		return nil, fmt.Errorf("not supported bits per sensor in measures")
	}

	return handler, nil
}

func GetSpecialDeviceForHost(hostId int) int32 {
	return int32(hostId) + MaxDeviceId
}

func GetHostForSpecialDevice(deviceId int32) (int, error) {
	if deviceId < MaxDeviceId {
		return 0, fmt.Errorf("not special device id")
	}
	return int(deviceId - MaxDeviceId), nil
}
