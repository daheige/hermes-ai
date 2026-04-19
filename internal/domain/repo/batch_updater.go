package repo

type BatchUpdater interface {
	Start()
	Stop()
	AddRecord(targetType int, id int, value int64)
}
