package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Load the Postgres drivers
)

type PostgresTransactionLogger struct {
	events chan<- Event
	errors <-chan error
	db     *sql.DB
}

type PostgresDBParams struct {
	dbName   string
	host     string
	user     string
	password string
}

func (l *PostgresTransactionLogger) verifyTableExists() (bool, error) {
	var result string
	const table = "transactions"

	rows, err := l.db.Query(fmt.Sprintf("SELECT to_regclass('public.%s');", table))
	defer rows.Close()
	if err != nil {
		return false, err
	}

	for rows.Next() && result != table {
		rows.Scan(&result)
	}

	return result == table, rows.Err()
}

func (l *PostgresTransactionLogger) createTable() error {
	var err error

	createQuery := `CREATE TABLE transaction (
						sequence 	BIGSERIAL PRIMARY KEY,
						event_type 	SMALLINT,
						key 		TEXT
						value 		TEXT
					);`
	_, err = l.db.Exec(createQuery)
	if err != nil {
		return err
	}

	return nil
}

func NewPostgresTrasnactionLogger(config PostgresDBParams) (TransactionLogger, error) {
	conStr := fmt.Sprintf("host=%s dbame=%s user=%s password=%s", config.host, config.dbName, config.user, config.password)

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	logger := &PostgresTransactionLogger{db: db}

	exists, err := logger.verifyTableExists()
	if err != nil {
		return nil, fmt.Errorf("failed to verify table exists: %w", err)
	}
	if !exists {
		if err = logger.createTable(); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return logger, nil
}

func (l *PostgresTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *PostgresTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *PostgresTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *PostgresTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors
	go func() {
		query := `INSERT INTO transactions 
							  (event_type, key, value) 
						 VALUES
						 	  ($1, $2, $3)
				 `

		for e := range events {
			_, err := l.db.Exec(query, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (l *PostgresTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		defer close(outEvent)
		defer close(outError)

		query := `SELECT sequence, event_type, key, value FROM transactions
                  ORDER BY sequence`

		rows, err := l.db.Query(query) // Run query; get result set
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}

		defer rows.Close() // This is important!

		e := Event{} // Create an empty Event

		for rows.Next() { // Iterate over the rows

			err = rows.Scan( // Read the values from the
				&e.Sequence, &e.EventType, // row into the Event.
				&e.Key, &e.Value)

			if err != nil {
				outError <- fmt.Errorf("error reading row: %w", err)
				return
			}

			outEvent <- e // Send e to the channel
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}
