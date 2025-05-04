package core

const (
	SystemUndefined32BitValue uint32 = 0x80000000
	SystemUndefined16BitValue uint16 = 0x8000
	UndefinedMeasureValue     uint32 = 0xFFFFFFFF
)

const (
	PackageFormatData                       byte = 0
	PackageFormatEvents                     byte = 1
	PackageFormatFullFailureStates          byte = 2
	PackageFormatFullAccidentStates         byte = 5
	PackageFormatFullObjectStates           byte = 6
	PackageFormatHeartbeat                  byte = 7
	PackageFormatChangeObjectStates         byte = 8
	PackageFormatChangeFailureStates        byte = 9
	PackageFormatChangeNotRespondingDevices byte = 10
)

const (
	PackageEventTypeFailureInfo                   byte = 1
	PackageEventTypeTimeMeasurement               byte = 2
	PackageEventTypeNoConnectionWithDevice        byte = 3
	PackageEventTypeFailurePrognosisAlgorithmInfo byte = 4
	PackageEventTypeNwaLeaveInfo                  byte = 5 // Выход из АНР
	PackageEventTypeNwaStateChangeInfo            byte = 6
	PackageEventTypeAccidentInfo                  byte = 7
	PackageEventTypeObjectState                   byte = 8
)
