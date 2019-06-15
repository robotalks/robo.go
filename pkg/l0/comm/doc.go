// Package comm provides L0 protocol support.
package comm

// L0 protocol is communicated between L0 firmware and L1 controller
// and focuses on robustness of data transferring that the communication
// is recoverable from errors over a peer-to-peer channel (e.g. serial port).
//
// This package uses a simple sequence based synchronization mechanism.
// It provides limited transfer error detection based on sequence
// check. However, it doesn't do any bit verification (e.g. CRC/Checksum)
// for simplicity and to be lightweighted.
// If needed, parity bits can be enabled on serial port for verification.
//
// Producer: L0 firmware
// Consumer: L1 controller
