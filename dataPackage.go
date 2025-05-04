package core

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"time"
)

type DataPackage struct {
	Time          uint64
	DeviceId      int32
	SensorCount   uint16
	BitsPerSensor byte
	Format        byte
	DataSize      uint16
	Data          []byte
}

func (data *DataPackage) Write(writer *bufio.Writer) {
	binary.Write(writer, binary.LittleEndian, data.Time)
	binary.Write(writer, binary.LittleEndian, data.DeviceId)
	binary.Write(writer, binary.LittleEndian, data.SensorCount)
	binary.Write(writer, binary.LittleEndian, data.BitsPerSensor)
	binary.Write(writer, binary.LittleEndian, data.Format)
	binary.Write(writer, binary.LittleEndian, data.DataSize)
	if data.DataSize > 0 {
		writer.Write(data.Data)
	}
}

func (data *DataPackage) Verify() error {
	return nil
}

func (data *DataPackage) GetPackageTime() time.Time {
	return GetTimeFromUnixMicroseconds(data.Time)
}

func (data *DataPackage) GetBase64String() string {
	return base64.StdEncoding.EncodeToString(data.Bytes())
}

func (data *DataPackage) Bytes() []byte {
	var replyBuf bytes.Buffer
	var writer = bufio.NewWriter(&replyBuf)
	data.Write(writer)
	writer.Flush()
	return replyBuf.Bytes()
}

func (data *DataPackage) IsCompressed() bool {
	if data.Format != PackageFormatData {
		return false
	}

	switch bitsPerSensor := data.BitsPerSensor; bitsPerSensor {
	case 2:
		var nExtraBytes uint16 = 0
		if data.SensorCount%4 != 0 {
			nExtraBytes = 1
		}
		return data.DataSize != data.SensorCount/4+nExtraBytes
	case 8:
		return data.DataSize != data.SensorCount
	case 16:
		return data.DataSize != data.SensorCount*2
	case 32:
		return data.DataSize != data.SensorCount*4
	default:
		return false
	}
}

func (data *DataPackage) String() string {
	return fmt.Sprintf("DevId=%d,Format=%d,Time=%s",
		data.DeviceId,
		data.Format,
		data.GetPackageTime().Format(time.RFC3339Nano))
}

func (data *DataPackage) Read(reader *bytes.Reader) error {

	var err error

	if err = binary.Read(reader, binary.LittleEndian, &data.Time); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &data.DeviceId); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &data.SensorCount); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &data.BitsPerSensor); err != nil {
		return err
	}
	if err = binary.Read(reader, binary.LittleEndian, &data.Format); err != nil {
		return err
	}

	if err = binary.Read(reader, binary.LittleEndian, &data.DataSize); err != nil {
		return err
	}

	if data.DataSize > 0 {
		data.Data = make([]byte, data.DataSize)
		// Читаем данные переменного размера
		var n int
		n, err = reader.Read(data.Data)

		if n != int(data.DataSize) {
			err = errors.New("error in data size")
		}
	}

	return err
}

func getObjectStateEvent(data []byte) (uint32, uint16) {
	objectId := binary.LittleEndian.Uint32(data)
	stateValue := binary.LittleEndian.Uint16(data[4:])

	return objectId, stateValue
}

type ObjectFailureEventInfo struct {
	IsStarted bool
	ObjectId  uint32
	FailureId uint32
	EventTime time.Time
}

type ObjectFpEventInfo struct {
	ObjectId    uint32
	AlgorithmId uint32
	StepIndex   int32
	EventTime   time.Time
}

/**
Выход или возвращение в САНР для объекта
*/
type ObjectNwaStateLeaveEventInfo struct {
	ObjectId    uint32
	AlgorithmId uint32
	StateId     int32 // состояние из которого выходим если IsStarted = true или в которое возвращаемся в противном случае
	IsStarted   bool  // true если выход из состояния
	EventTime   time.Time
}

type ObjectAccidentEventInfo struct {
	ObjectId     uint32    // ид. объекта
	AccidentType byte      // 1 выход из АНР, 2 - срабатывание АП
	AlgorithmId  int32     // -1 если тип 1, или ид. АП
	StartTime    time.Time // время начала инцидента
	EndTime      time.Time // время завершения инцидента (или все 0, если не завершен)
}

/**
Изменение САНР для объекта
*/
type ObjectNwaStateChangeEventInfo struct {
	ObjectId   uint32
	NwaStateId int32
	EventTime  time.Time
}

type ObjectFailureKey struct {
	ObjectId  uint32
	FailureId uint32
}

type ObjectAccidentKey struct {
	ObjectId   uint32
	AccidentId int32
}

func getObjectFailureEvent(data []byte) *ObjectFailureEventInfo {
	var result = &ObjectFailureEventInfo{}

	result.ObjectId = binary.LittleEndian.Uint32(data)
	result.FailureId = binary.LittleEndian.Uint32(data[4:])

	if data[8] == 0 {
		result.IsStarted = false
	} else {
		result.IsStarted = true
	}
	// В пакете время в микросекундах
	var t = binary.LittleEndian.Uint64(data[9:])
	// Переводим в наносекунды
	result.EventTime = GetTimeFromUnixMicroseconds(t)

	return result
}

func getObjectFpEvent(data []byte) *ObjectFpEventInfo {
	var result = &ObjectFpEventInfo{}

	result.AlgorithmId = binary.LittleEndian.Uint32(data)
	result.ObjectId = binary.LittleEndian.Uint32(data[4:])
	result.StepIndex = int32(binary.LittleEndian.Uint32(data[8:]))

	// В пакете время в микросекундах
	var t = binary.LittleEndian.Uint64(data[12:])
	// Переводим в наносекунды
	result.EventTime = GetTimeFromUnixMicroseconds(t)

	return result
}

func getNwaChangeStateEvent(data []byte) (uint32, int32) {
	objectId := binary.LittleEndian.Uint32(data)
	stateId := int32(binary.LittleEndian.Uint32(data[4:]))
	return objectId, stateId
}

func getObjectNwaLeaveEvent(data []byte) *ObjectNwaStateLeaveEventInfo {
	var result = &ObjectNwaStateLeaveEventInfo{}

	result.ObjectId = binary.LittleEndian.Uint32(data)
	result.AlgorithmId = binary.LittleEndian.Uint32(data[4:])
	// САНР из которого вышел или в которое вернулся объект
	result.StateId = int32(binary.LittleEndian.Uint32(data[8:]))

	if data[12] == 0 {
		// Объект вошел в САНР
		result.IsStarted = false
	} else {
		// Объект вышел из САНР
		result.IsStarted = true
	}
	// В пакете время в микросекундах
	var t = binary.LittleEndian.Uint64(data[13:])
	// Переводим в наносекунды
	result.EventTime = GetTimeFromUnixMicroseconds(t)

	return result
}

func getObjectAccidentEvent(data []byte) *ObjectAccidentEventInfo {
	accidentType := data[0]
	var algorithmId int32
	binary.Read(bytes.NewReader(data[1:]), binary.LittleEndian, &algorithmId)
	objectId := binary.LittleEndian.Uint32(data[5:])
	startTime := binary.LittleEndian.Uint64(data[9:])
	endTime := binary.LittleEndian.Uint64(data[17:])

	return &ObjectAccidentEventInfo{
		AccidentType: accidentType,
		AlgorithmId:  algorithmId,
		ObjectId:     objectId,
		StartTime:    GetTimeFromUnixMicroseconds(startTime),
		EndTime:      GetTimeFromUnixMicroseconds(endTime)}
}

func (data *DataPackage) ParseFullObjectStatePackage() (map[uint32]uint16, error) {

	if data.Format != PackageFormatFullObjectStates {
		return nil, fmt.Errorf("expected full state package format")
	}

	// Размер данных должен быть кратен 7 (маркер byte, идентификатор объекта int, код состояния short)
	if len(data.Data)%7 != 0 {
		return nil, fmt.Errorf("events data size should be 7 * nItems")
	}

	var objectStates = make(map[uint32]uint16)

	for i := 0; i < len(data.Data)/7; i++ {
		var curPos = i * 7
		var marker = data.Data[curPos]

		if marker != PackageEventTypeObjectState {
			return nil, fmt.Errorf("unexpected marker in message %d", marker)
		}

		objectId, objectState := getObjectStateEvent(data.Data[curPos+1:])

		objectStates[objectId] = objectState
	}

	return objectStates, nil
}

func (data *DataPackage) ParseFullFailureStatePackage() (map[ObjectFailureKey]*ObjectFailureEventInfo, error) {

	if data.Format != PackageFormatFullFailureStates {
		return nil, fmt.Errorf("expected full failure package format")
	}

	if len(data.Data)%18 != 0 {
		return nil, fmt.Errorf("failure full state message.events data size should be 18 * nItems")
	}

	var objectFailuresFullState = make(map[ObjectFailureKey]*ObjectFailureEventInfo)

	for i := 0; i < len(data.Data)/18; i++ {
		var curPos = i * 18
		var marker = data.Data[curPos]

		if marker != PackageEventTypeFailureInfo {
			return nil, fmt.Errorf("unexpected marker %d in failure full state message", marker)
		}

		failureEvent := getObjectFailureEvent(data.Data[curPos+1:])

		objectFailuresFullState[ObjectFailureKey{failureEvent.ObjectId, failureEvent.FailureId}] = failureEvent
	}

	return objectFailuresFullState, nil
}

type PackageEvents struct {
	ObjectStates               map[uint32]uint16
	ObjectFailuresChangeState  map[ObjectFailureKey]*ObjectFailureEventInfo
	ObjectAccidentsChangeState map[ObjectAccidentKey]*ObjectAccidentEventInfo
	ObjectFpChangeState        map[uint32]*ObjectFpEventInfo
	ObjectNwaChangeState       map[uint32]*ObjectNwaStateLeaveEventInfo // Переход объекта в САНР или выход из АНР
	ObjectNwaStateLeaveEnter   map[uint32]*ObjectNwaStateChangeEventInfo
}

func (events *PackageEvents) GetObjects() mapset.Set {
	result := mapset.NewSet()

	for k := range events.ObjectStates {
		result.Add(int(k))
	}
	for k := range events.ObjectFpChangeState {
		result.Add(int(k))
	}

	for k := range events.ObjectNwaChangeState {
		result.Add(int(k))
	}

	for k := range events.ObjectNwaStateLeaveEnter {
		result.Add(int(k))
	}

	for k := range events.ObjectFailuresChangeState {
		result.Add(int(k.ObjectId))
	}

	for k := range events.ObjectAccidentsChangeState {
		result.Add(int(k.ObjectId))
	}

	return result
}

func getEventRecordSize(marker byte) (int, error) {
	switch marker {
	case PackageEventTypeFailureInfo:
		return 18, nil
	case PackageEventTypeTimeMeasurement:
		return 13, nil
	case PackageEventTypeNoConnectionWithDevice:
		return 18, nil
	case PackageEventTypeFailurePrognosisAlgorithmInfo:
		return 21, nil
	case PackageEventTypeNwaLeaveInfo:
		return 22, nil
	case PackageEventTypeNwaStateChangeInfo:
		return 13, nil // minimal size if there are no states in the list
	case PackageEventTypeAccidentInfo:
		return 26, nil
	case PackageEventTypeObjectState:
		return 7, nil
	default:
		return 0, fmt.Errorf("unknown marker %d", marker)
	}
}

func (data *DataPackage) ParseEventsPackage() (*PackageEvents, error) {

	if !(data.Format == PackageFormatEvents ||
		data.Format == PackageFormatChangeObjectStates ||
		data.Format == PackageFormatChangeFailureStates) {
		return nil, fmt.Errorf("expected events package format")
	}

	addChangeObjectStateEventsOnly := data.Format == PackageEventTypeObjectState

	result := &PackageEvents{
		ObjectStates:               make(map[uint32]uint16),
		ObjectFailuresChangeState:  make(map[ObjectFailureKey]*ObjectFailureEventInfo),
		ObjectAccidentsChangeState: make(map[ObjectAccidentKey]*ObjectAccidentEventInfo),
		ObjectFpChangeState:        make(map[uint32]*ObjectFpEventInfo),
		ObjectNwaChangeState:       make(map[uint32]*ObjectNwaStateLeaveEventInfo),
		ObjectNwaStateLeaveEnter:   make(map[uint32]*ObjectNwaStateChangeEventInfo)}

	// Отфильтровываем пакеты об изменении состояния объекта
	for i := 0; i < len(data.Data); {
		marker := data.Data[i]

		size, err := getEventRecordSize(marker)
		if err != nil {
			return nil, err
		}

		if i+size > len(data.Data) {
			return nil, fmt.Errorf("incorrect package size")
		}

		switch marker {

		case PackageEventTypeFailureInfo:
			if !addChangeObjectStateEventsOnly {
				failureEvent := getObjectFailureEvent(data.Data[i+1:])
				result.ObjectFailuresChangeState[ObjectFailureKey{
					ObjectId:  failureEvent.ObjectId,
					FailureId: failureEvent.FailureId}] = failureEvent
			}

		case PackageEventTypeAccidentInfo:
			if !addChangeObjectStateEventsOnly {
				accidentEvent := getObjectAccidentEvent(data.Data[i+1:])
				result.ObjectAccidentsChangeState[ObjectAccidentKey{ObjectId: accidentEvent.ObjectId,
					AccidentId: accidentEvent.AlgorithmId}] = accidentEvent
			}

		case PackageEventTypeFailurePrognosisAlgorithmInfo:
			if !addChangeObjectStateEventsOnly {
				fpEvent := getObjectFpEvent(data.Data[i+1:])
				result.ObjectFpChangeState[fpEvent.ObjectId] = fpEvent
			}

		case PackageEventTypeNwaLeaveInfo:
			if !addChangeObjectStateEventsOnly {
				nwaEvent := getObjectNwaLeaveEvent(data.Data[i+1:])
				result.ObjectNwaChangeState[nwaEvent.ObjectId] = nwaEvent
			}

		case PackageEventTypeNwaStateChangeInfo:
			nwaStateEventTime := GetTimeFromUnixMicroseconds(binary.LittleEndian.Uint64(data.Data[i+1:]))

			nObjectStates := binary.LittleEndian.Uint32(data.Data[i+9:])
			size += int(nObjectStates) * 8
			if i+size > len(data.Data) {
				return nil, fmt.Errorf("incorrect size for nwa event package")
			}

			if !addChangeObjectStateEventsOnly {
				var i uint32 = 0
				for i = 0; i < nObjectStates; i++ {
					oId, stateId := getNwaChangeStateEvent(data.Data[i*8+13:])
					result.ObjectNwaStateLeaveEnter[oId] = &ObjectNwaStateChangeEventInfo{
						EventTime: nwaStateEventTime,
						ObjectId:  oId, NwaStateId: stateId}
				}
			}

		case PackageEventTypeObjectState:
			oId, v := getObjectStateEvent(data.Data[i+1:])
			result.ObjectStates[oId] = v
		}

		i += size
	}

	return result, nil
}
