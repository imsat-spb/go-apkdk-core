package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type NetworkPackage struct {
	HostId    int32
	PackageId int32
	Data      DataPackage
}

func (res *NetworkPackage) String() string {
	return fmt.Sprintf("HostId= %d, PackageId=%d, Content=[%s]", res.HostId, res.PackageId, &res.Data)
}

func (res *NetworkPackage) Read(reader *bytes.Reader) error {
	var err error

	if err = binary.Read(reader, binary.LittleEndian, &res.HostId); err != nil {
		return err
	}

	if err = binary.Read(reader, binary.LittleEndian, &res.PackageId); err != nil {
		return err
	}

	res.Data = DataPackage{}
	return res.Data.Read(reader)
}
