package keysstorage

func safeClose(obj Closable) {
	if obj != nil {
		obj.Close()
	}
}
