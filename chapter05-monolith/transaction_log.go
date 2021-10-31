package main

type EventType byte

const (
	_                     = iota // iota = 0; bu datayı görmezden gel
	EventDelete EventType = iota // iota = 1
	EventPut                     // burada iota otomatik olarak devereye giriyor
	// önceki tanımı alıp oradan devam ediyor, haliyle
	// EventPut'un tipi EventType oluyor, değeri ise
	// otomatik olarak bir önceki değeri 1 fazlası oluyor

)

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error

	ReadEvents() (<-chan Event, <-chan error)

	Run()
}
