package main

import (
	"bufio"
	"fmt"
	"os"
)

type FileTransactionLogger struct {
	events       chan<- Event // Write-only channel for sending events
	errors       <-chan error // Read-only channel for receiving errors
	lastSequence uint64       // The last used event sequence number
	file         *os.File     // The location of the transaction log
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

// Burada **dönüş tiplerine** dikkat!
// Fonksiyon `TransactionLogger` `interface`inin döndüğünü gösteriyor
// fakat bunun yerine bu `interface`'i destekleyen öğenin `pointer`ı
// dönüyor.
//
// Bunun sebebi `GO`'nun `interface`lere `pointer`'ı desteklememesi,
// fakat bu `interface`i implemente etmiş tiplere `pointer`ı desteklemesi
func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	// `TransactionLogger` `interface`'i yerine bu `interface`in yerine
	// `pointer`ı dönüyor
	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) Run() {
	// `events` `channel`ı oluşturulurken ona bir buffer tanımlanıyor. Böylece
	// buffer dolana kadar sistemişn bloke olması engellenmiş oluyor. Ancak buffer
	// doluysa bu `channel`ı kullanan fonksiyonlar (örneğin `WritePut` ve `WriteDelete`)
	// fonksiyonları bloke edilecek ve mevcut `go routin`ler tamamalana kadar bbloke
	// edilecekler
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++

			_, err := fmt.Fprintf(l.file, "%d\t%d\t%s\t%s\n", l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file) // Create a Scanner for l.file
	outEvent := make(chan Event)        // An unbuffered Event channel
	outError := make(chan error, 1)     // A buffered error channel

	go func() {
		var e Event

		defer func() {
			close(outEvent) // Close the channels when the
		}()
		defer close(outError) // goroutine ends

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s",
				&e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {

				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			// Sanity check! Are the sequence numbers in increasing order?
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = e.Sequence // Update last used sequence #

			outEvent <- e // Send the event along
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}
