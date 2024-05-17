package dbclient

func CheckMediumBlob(dt []byte) bool {
	if len(dt) <= ((1 << 24) - 1) {
		return true
	}

	return false
}
