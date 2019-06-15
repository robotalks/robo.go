// +build linux

package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type device struct {
	file        *os.File
	index       int
	name        string
	axisCount   uint8
	buttonCount uint8
}

// Open opens the device with specified index.
func Open(index int) (Device, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/input/js%d", index), os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	d := &device{file: f, index: index}

	errno := d.ioctl(iocGAXES, unsafe.Pointer(&d.axisCount))
	if errno == 0 {
		errno = d.ioctl(iocGBUTTONS, unsafe.Pointer(&d.buttonCount))
	}
	if errno == 0 {
		var buf [256]byte
		errno = d.ioctl(iocGNAME, unsafe.Pointer(&buf))
		if errno == 0 {
			if pos := bytes.IndexByte(buf[:], 0); pos >= 0 {
				d.name = string(buf[:pos])
			} else {
				d.name = string(buf[:])
			}
		}
	}
	if errno != 0 {
		d.file.Close()
		return nil, errno
	}
	return d, nil
}

// DetectAndOpen detects a next available device from startIndex and opens it.
func DetectAndOpen(startIndex int) (Device, error) {
	for index := startIndex; index < 256; index++ {
		d, err := Open(index)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		return d, nil
	}
	return nil, nil
}

// Close implements Device.
func (d *device) Close() error {
	return d.file.Close()
}

// Index implements Device.
func (d *device) Index() int {
	return d.index
}

// Name implements Device.
func (d *device) Name() string {
	return d.name
}

// AxisCount implements Device.
func (d *device) AxisCount() int {
	return int(d.axisCount)
}

// ButtonCount implements Device.
func (d *device) ButtonCount() int {
	return int(d.buttonCount)
}

// ReadEvent implements Device.
func (d *device) ReadEvent() (Event, error) {
	buf := make([]byte, 8)
	if _, err := d.file.Read(buf); err != nil {
		return nil, err
	}
	var ev event
	if err := binary.Read(bytes.NewReader(buf), binary.LittleEndian, &ev); err != nil {
		return nil, err
	}
	switch ev.Type & (evBTN | evAXIS) {
	case evBTN:
		return &buttonEvent{event: ev}, nil
	case evAXIS:
		return &axisEvent{event: ev}, nil
	}
	return &ev, nil
}

type event struct {
	Time   uint32
	Value  int16
	Type   uint8
	Number uint8
}

func (e *event) IsInit() bool {
	return e.Type&evINIT != 0
}

func (e *event) Index() int {
	return int(e.Number)
}

type axisEvent struct {
	event
}

func (e *axisEvent) Value() int {
	return int(e.event.Value)
}

type buttonEvent struct {
	event
}

func (e *buttonEvent) Pressed() bool {
	return e.Value != 0
}

const (
	iocGAXES    uint = 0x80016a11
	iocGBUTTONS uint = 0x80016a12
	iocGNAME    uint = 0x80ff6a13

	evINIT uint8 = 0x80
	evBTN  uint8 = 0x01
	evAXIS uint8 = 0x02
)

func (d *device) ioctl(req uint, ptr unsafe.Pointer) syscall.Errno {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(d.file.Fd()), uintptr(req), uintptr(ptr))
	return err
}
