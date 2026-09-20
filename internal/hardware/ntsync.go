package hardware

import "os"

// HasNTSync returns true if the /dev/ntsync kernel fast synchronization device node exists.
// NTSync provides low-overhead in-kernel synchronization for Wine/Proton games,
// replacing legacy eventfd/futex (fsync/esync) synchronization mechanisms.
func HasNTSync() bool {
	info, err := os.Stat("/dev/ntsync")
	if err != nil {
		return false
	}
	// Verify it is a character device
	return info.Mode()&os.ModeDevice != 0 && info.Mode()&os.ModeCharDevice != 0
}
