package errors

import (
	"fmt"
)

type ConfigFileInitializeError struct {
	basicAliothError
	filePath  string
	operation string
	err       error
}

func (e *ConfigFileInitializeError) Error() string {
	return fmt.Errorf("error occurred when %s config file [%s]: %w", e.filePath, e.operation, e.err).Error()
}

func NewOpenConfigFileInitializeError(filePath string, err error) AliothError {
	return &ConfigFileInitializeError{
		filePath:  filePath,
		operation: "open",
		err:       err,
	}
}

func NewReadConfigFileInitializeError(filePath string, err error) AliothError {
	return &ConfigFileInitializeError{
		filePath:  filePath,
		operation: "read",
		err:       err,
	}
}

type LogFileExecutingError struct {
	basicAliothError
	filePath  string
	operation string
	err       error
}

func (e *LogFileExecutingError) Error() string {
	return fmt.Errorf("error occurred when %s log file [%s]: %w", e.filePath, e.operation, e.err).Error()
}

func NewOpenLogFileError(filePath string, err error) AliothError {
	return &LogFileExecutingError{
		filePath:  filePath,
		operation: "open",
		err:       err,
	}
}

func NewReadLogFileError(filePath string, err error) AliothError {
	return &LogFileExecutingError{
		filePath:  filePath,
		operation: "read",
		err:       err,
	}
}

func NewWriteLogFileError(filePath string, err error) AliothError {
	return &LogFileExecutingError{
		filePath:  filePath,
		operation: "write",
		err:       err,
	}
}

func NewCloseLogFileError(filePath string, err error) AliothError {
	return &LogFileExecutingError{
		filePath:  filePath,
		operation: "close",
		err:       err,
	}
}

type DatabaseInitializeError struct {
	basicAliothError
	databaseHost string
	databasePort int
	databaseName string
	databaseUser string
	err          error
}

func (e *DatabaseInitializeError) Error() string {
	return fmt.Errorf("error occurred when initialize database [%s:%d@%s/%s]: %w",
		e.databaseUser, e.databasePort, e.databaseHost, e.databaseName, e.err).Error()
}

func NewDatabaseInitializeError(databaseHost string, databasePort int, databaseName, databaseUser string, err error) AliothError {
	return &DatabaseInitializeError{
		databaseHost: databaseHost,
		databasePort: databasePort,
		databaseName: databaseName,
		databaseUser: databaseUser,
		err:          err,
	}
}

type DatabaseSyncModelsError struct {
	basicAliothError
	models []any
	err    error
}

func (e *DatabaseSyncModelsError) Error() string {
	return fmt.Errorf("error occurred when sync database models [%v]: %w", e.models, e.err).Error()
}

func NewDatabaseSyncModelsError(models []any, err error) AliothError {
	return &DatabaseSyncModelsError{
		models: models,
		err:    err,
	}
}
