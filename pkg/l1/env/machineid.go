package env

import (
	"github.com/denisbrodbeck/machineid"
)

// MachineID retrieves the unique ID identifying the machine.
func MachineID() string {
	id, err := machineid.ID()
	if err != nil {
		panic(err)
	}
	return id
}
