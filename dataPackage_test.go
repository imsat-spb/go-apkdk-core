package core

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestIsCompressed(t *testing.T) {
	tests := []struct {
		name           string
		data           *DataPackage
		expectedResult bool
	}{
		{"EmptyPackage", &DataPackage{}, false},
		{"Package1", &DataPackage{
			Format:        PackageFormatHeartbeat,
			BitsPerSensor: 8,
			SensorCount:   1,
			DataSize:      1,
			Data:          []byte{0}}, false},
		{"Package2", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 32,
			SensorCount:   1,
			DataSize:      4,
			Data:          []byte{0, 0, 0, 0}}, false},
		{"Package3", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 32,
			SensorCount:   10,
			DataSize:      4,
			Data:          []byte{0, 0, 0, 0}}, true},
		{"Package4", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 8,
			SensorCount:   2,
			DataSize:      1,
			Data:          []byte{1}}, true},
		{"Package5", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 8,
			SensorCount:   1,
			DataSize:      1,
			Data:          []byte{1}}, false},
		{"Package6", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 16,
			SensorCount:   2,
			DataSize:      2,
			Data:          []byte{0, 0}}, true},
		{"Package7", &DataPackage{
			Format:        PackageFormatData,
			BitsPerSensor: 16,
			SensorCount:   1,
			DataSize:      2,
			Data:          []byte{1}}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.data.IsCompressed(), test.expectedResult)
		})
	}
}

func TestGetObjectStateEvent(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		objectId uint32
		stateId  uint16
	}{
		{"Package1", []byte{100, 0, 0, 0, 200, 0}, 100, 200},
		{"Package2", []byte{2, 0, 0, 0, 1, 0}, 2, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			objectId, stateId := getObjectStateEvent(test.data)
			assert.Equal(t, test.objectId, objectId)
			assert.Equal(t, test.stateId, stateId)
		})
	}
}

func TestGetEventRecordSize(t *testing.T) {
	tests := []struct {
		name    string
		marker  byte
		size    int
		isValid bool
	}{
		{"test1", PackageEventTypeFailureInfo, 18, true},
		{"test2", PackageEventTypeTimeMeasurement, 13, true},
		{"test3", PackageEventTypeNoConnectionWithDevice, 18, true},
		{"test4", PackageEventTypeFailurePrognosisAlgorithmInfo, 21, true},
		{"test5", PackageEventTypeNwaLeaveInfo, 22, true},
		{"test6", PackageEventTypeNwaStateChangeInfo, 13, true},
		{"test7", PackageEventTypeAccidentInfo, 26, true},
		{"test8", PackageEventTypeObjectState, 7, true},
		{"test9", 0, 0, false},
		{"test10", 9, 0, false},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			size, err := getEventRecordSize(test.marker)

			if !test.isValid {
				assert.Zero(t, size)
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, test.size, size)
				assert.Nil(t, err)
			}
		})
	}
}

func TestParseFullObjectStatesPackage(t *testing.T) {
	tests := []struct {
		name    string
		data    *DataPackage
		result  map[uint32]uint16
		isValid bool
	}{
		{"test1", &DataPackage{Format: PackageFormatData}, nil, false},
		{"test2", &DataPackage{Format: PackageFormatFullObjectStates,
			BitsPerSensor: 8, DataSize: 7, SensorCount: 7, Data: []byte{PackageEventTypeObjectState, 100, 0, 0, 0, 1}},
			nil, false},
		{"test3", &DataPackage{Format: PackageFormatFullObjectStates,
			BitsPerSensor: 8, DataSize: 7, SensorCount: 7, Data: []byte{PackageEventTypeAccidentInfo, 100, 0, 0, 0, 1, 0}},
			nil, false},
		{"test4", &DataPackage{Format: PackageFormatFullObjectStates,
			BitsPerSensor: 8, DataSize: 7, SensorCount: 7, Data: []byte{PackageEventTypeObjectState, 100, 0, 0, 0, 1, 0}},
			map[uint32]uint16{100: 1}, true},
		{"test5", &DataPackage{Format: PackageFormatFullObjectStates,
			BitsPerSensor: 8, DataSize: 14, SensorCount: 14, Data: []byte{
				PackageEventTypeObjectState, 100, 0, 0, 0, 1, 0,
				PackageEventTypeObjectState, 200, 0, 0, 0, 1, 0}},
			map[uint32]uint16{100: 1, 200: 1}, true},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result, err := test.data.ParseFullObjectStatePackage()

			if !test.isValid {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(test.result, result))
			}
		})
	}
}

func TestParseFullFailureStatePackage(t *testing.T) {
	timeUnix := GetUnixMicrosecondsFromTime(time.Now())
	dataTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataTime, uint64(timeUnix))

	test4StartTime := GetTimeFromUnixMicroseconds(binary.LittleEndian.Uint64(dataTime))
	tests := []struct {
		name    string
		data    *DataPackage
		result  map[ObjectFailureKey]*ObjectFailureEventInfo
		isValid bool
	}{
		{"test1", &DataPackage{Format: PackageFormatData}, nil, false},
		{"test2", &DataPackage{Format: PackageFormatFullFailureStates,
			BitsPerSensor: 8, DataSize: 18, SensorCount: 18, Data: append([]byte{
				PackageEventTypeFailureInfo, 100, 0, 0, 0, 1, 0, 0, 0}, dataTime...)},
			nil, false},
		/*
			{"test3", &DataPackage{Format:PackageFormatFullObjectStates,
				BitsPerSensor:8, DataSize:7,SensorCount:7, Data:[]byte{PackageEventTypeAccidentInfo,100,0,0,0,1,0}}, false},*/
		{"test4", &DataPackage{Format: PackageFormatFullFailureStates,
			BitsPerSensor: 8, DataSize: 18, SensorCount: 18, Data: append([]byte{
				PackageEventTypeFailureInfo, 100, 0, 0, 0, 1, 0, 0, 0, 1}, dataTime...)},
			map[ObjectFailureKey]*ObjectFailureEventInfo{
				{ObjectId: 100, FailureId: 1}: {ObjectId: 100, FailureId: 1, IsStarted: true, EventTime: test4StartTime}},
			true},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result, err := test.data.ParseFullFailureStatePackage()

			if !test.isValid {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(test.result, result))
			}
		})
	}
}

func getSliceFromInt32(value int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, value)
	return buf.Bytes()
}

func getTimeAndSlice(timeValue time.Time) (time.Time, []byte) {
	timeUnix := GetUnixMicrosecondsFromTime(timeValue) // В пакете время в микросекундах
	dataTime := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataTime, uint64(timeUnix))

	time := GetTimeFromUnixMicroseconds(binary.LittleEndian.Uint64(dataTime))
	return time, dataTime
}

func TestParseEventsWithAccidentsPackage(t *testing.T) {
	timeNow := time.Now()
	testStartTime, testTimeSlice := getTimeAndSlice(timeNow)

	algId := getSliceFromInt32(-1)

	testData := append([]byte{
		PackageEventTypeAccidentInfo, // marker
		1},                           // event type (leave anr)
		algId...) // algorithm id (-1!!!)
	testData = append(testData, []byte{
		200, 0, 0, 0, // object id
	}...)

	testData = append(testData, testTimeSlice...) // start time

	testData = append(testData, testTimeSlice...)

	tests := []struct {
		name    string
		data    *DataPackage
		result  *PackageEvents
		isValid bool
	}{
		{name: "test1", data: &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 26, DataSize: 26,
			Data: testData,
		},
			result: &PackageEvents{ObjectStates: make(map[uint32]uint16),
				ObjectFailuresChangeState: make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState:      make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: map[ObjectAccidentKey]*ObjectAccidentEventInfo{
					{ObjectId: 200, AccidentId: -1}: { // TODO: -1
						ObjectId:     200,
						AccidentType: 1,
						AlgorithmId:  -1,
						StartTime:    testStartTime,
						EndTime:      testStartTime,
					},
				},
				ObjectFpChangeState:      make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter: make(map[uint32]*ObjectNwaStateChangeEventInfo)}, isValid: true},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result, err := test.data.ParseEventsPackage()

			if !test.isValid {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(test.result, result))
			}
		})
	}
}

func TestReadPackage(t *testing.T) {
	timeNow := time.Now()
	testStartTime, testTimeSlice := getTimeAndSlice(timeNow)

	_ = testStartTime

	testData := append(testTimeSlice,
		[]byte{
			0, 0, 0, 0, // special device id
			1, 0, // sensor count
			8, //  bit per sensor
			PackageFormatHeartbeat,
			1, 0, // data size,
			1, // data
		}...)

	var dataPackage DataPackage

	r := bytes.NewReader(testData)
	dataPackage.Read(r)

	assert.Equal(t, PackageFormatHeartbeat, dataPackage.Format)
	assert.Equal(t, uint16(1), dataPackage.DataSize)
	assert.Equal(t, uint8(8), dataPackage.BitsPerSensor)
	assert.Equal(t, uint16(1), dataPackage.SensorCount)
}

func TestParseEventsPackage(t *testing.T) {
	timeNow := time.Now()
	test5StartTime, test5TimeSlice := getTimeAndSlice(timeNow)

	test5Data := append([]byte{PackageEventTypeFailureInfo, 100, 0, 0, 0, 1, 0, 0, 0, 1}, test5TimeSlice...)

	time15SecondsBefore := time.Now().Add(-time.Second * 15)
	test2_100_StartTime, test2_100_TimeSlice := getTimeAndSlice(time15SecondsBefore)

	time30SecondsBefore := time.Now().Add(-time.Second * 30)

	test2_200_StartTime, test2_200_TimeSlice := getTimeAndSlice(time30SecondsBefore)

	test2_100Data := append(
		[]byte{PackageEventTypeNwaLeaveInfo, 100, 0, 0, 0, 1, 0, 0, 0, 121, 0, 0, 0, 1},
		test2_100_TimeSlice...)

	test2_200Data := append(
		[]byte{PackageEventTypeNwaLeaveInfo, 200, 0, 0, 0, 4, 0, 0, 0, 133, 0, 0, 0, 0},
		test2_200_TimeSlice...)

	test2Data := append(test2_100Data, test2_200Data...)

	timeOneSecondBefore := timeNow.Add(-time.Second)
	test3_100_StartTime, test3_100_TimeSlice := getTimeAndSlice(timeOneSecondBefore)

	test3Data := append(
		[]byte{PackageEventTypeFailurePrognosisAlgorithmInfo, 4, 0, 0, 0, 100, 0, 0, 0, 67, 0, 0, 0}, test3_100_TimeSlice...)

	test6StartTime := test3_100_StartTime
	test6_100Data := []byte{100, 0, 0, 0, 23, 0, 0, 0}
	test6_200Data := []byte{200, 0, 0, 0, 255, 255, 255, 255}

	test6Data := append([]byte{PackageEventTypeNwaStateChangeInfo}, test3_100_TimeSlice...)
	test6Data = append(test6Data, []byte{2, 0, 0, 0}...)
	test6Data = append(test6Data, test6_100Data...)
	test6Data = append(test6Data, test6_200Data...)

	test8StartTime := test6StartTime
	test8TimeSlice := test3_100_TimeSlice

	tests := []struct {
		name    string
		data    *DataPackage
		result  *PackageEvents
		isValid bool
	}{
		{"test1", &DataPackage{Format: PackageFormatData}, nil, false},

		{"test2", &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 42, DataSize: 42,
			Data: test2Data,
		},
			&PackageEvents{ObjectStates: make(map[uint32]uint16),
				ObjectFailuresChangeState: make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState: map[uint32]*ObjectNwaStateLeaveEventInfo{
					100: {IsStarted: true, ObjectId: 100, StateId: 121, AlgorithmId: 1, EventTime: test2_100_StartTime},
					200: {IsStarted: false, ObjectId: 200, StateId: 133, AlgorithmId: 4, EventTime: test2_200_StartTime},
				},
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)}, true},

		{"test3", &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 21, DataSize: 21,
			Data: test3Data,
		},
			&PackageEvents{ObjectStates: make(map[uint32]uint16),
				ObjectFailuresChangeState:  make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState: map[uint32]*ObjectFpEventInfo{
					100: {ObjectId: 100, AlgorithmId: 4, StepIndex: 67, EventTime: test3_100_StartTime},
				},

				ObjectNwaStateLeaveEnter: make(map[uint32]*ObjectNwaStateChangeEventInfo)}, true},

		{"test4", &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 14, DataSize: 14, Data: []byte{
			PackageEventTypeObjectState, 100, 0, 0, 0, 1, 0,
			PackageEventTypeObjectState, 200, 0, 0, 0, 1, 0,
		}},
			&PackageEvents{ObjectStates: map[uint32]uint16{100: 1, 200: 1},
				ObjectFailuresChangeState:  make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)}, true},

		{"test5", &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 32, DataSize: 32,
			Data: append(append([]byte{
				PackageEventTypeObjectState, 100, 0, 0, 0, 1, 0}, test5Data...),
				PackageEventTypeObjectState, 200, 0, 0, 0, 1, 0),
		},
			&PackageEvents{ObjectStates: map[uint32]uint16{100: 1, 200: 1},
				ObjectFailuresChangeState: map[ObjectFailureKey]*ObjectFailureEventInfo{
					{ObjectId: 100, FailureId: 1}: {ObjectId: 100, FailureId: 1, IsStarted: true, EventTime: test5StartTime}},
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)},
			true},

		{"test6", &DataPackage{Format: PackageFormatEvents, BitsPerSensor: 8, SensorCount: 29, DataSize: 29,
			Data: test6Data,
		},
			&PackageEvents{ObjectStates: make(map[uint32]uint16),
				ObjectFailuresChangeState:  make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter: map[uint32]*ObjectNwaStateChangeEventInfo{
					100: {ObjectId: 100, NwaStateId: 23, EventTime: test6StartTime},
					200: {ObjectId: 200, NwaStateId: -1, EventTime: test6StartTime},
				},
			},
			true},
		{"test7", &DataPackage{Format: PackageFormatChangeObjectStates, BitsPerSensor: 8, SensorCount: 14, DataSize: 14, Data: []byte{
			PackageEventTypeObjectState, 100, 0, 0, 0, 1, 0,
			PackageEventTypeObjectState, 200, 0, 0, 0, 1, 0,
		}},
			&PackageEvents{ObjectStates: map[uint32]uint16{100: 1, 200: 1},
				ObjectFailuresChangeState:  make(map[ObjectFailureKey]*ObjectFailureEventInfo),
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)}, true},

		{"test8", &DataPackage{Format: PackageFormatChangeFailureStates, BitsPerSensor: 8, SensorCount: 18, DataSize: 18,
			Data: append([]byte{
				PackageEventTypeFailureInfo, 100, 0, 0, 0, 1, 0, 0, 0, 1,
			}, test8TimeSlice...)},
			&PackageEvents{ObjectStates: make(map[uint32]uint16),
				ObjectFailuresChangeState: map[ObjectFailureKey]*ObjectFailureEventInfo{
					{ObjectId: 100, FailureId: 1}: {ObjectId: 100, FailureId: 1, IsStarted: true, EventTime: test8StartTime},
				},
				ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
				ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
				ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
				ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)}, true},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result, err := test.data.ParseEventsPackage()

			if !test.isValid {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.True(t, reflect.DeepEqual(test.result, result))
			}
		})
	}
}

func TestGetObjectsFromEvents(t *testing.T) {

	now := time.Now()

	tests := []struct {
		name     string
		data     PackageEvents
		expected []int
	}{
		{"test1", PackageEvents{ObjectStates: map[uint32]uint16{
			1000: 1,
			2000: 3,
		}}, []int{1000, 2000}},

		{"test2", PackageEvents{ObjectStates: map[uint32]uint16{
			1000: 1,
			2000: 3,
		}, ObjectNwaChangeState: map[uint32]*ObjectNwaStateLeaveEventInfo{
			1000: {ObjectId: 1000, StateId: 10, IsStarted: true, EventTime: now},
			3000: {ObjectId: 3000, StateId: 11, IsStarted: false, EventTime: now},
		}}, []int{1000, 3000, 2000}},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			result := test.data.GetObjects()

			for _, i := range test.expected {
				assert.True(t, result.Contains(i))
			}

			assert.Equal(t, result.Cardinality(), len(test.expected))
			result.Contains()

		})
	}
}
