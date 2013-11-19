package db

type databaseFilter interface {
	Trigger(event string, args... interface {}) ([]interface {}, error)
}

