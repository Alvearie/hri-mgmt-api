/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logwrapper

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

const (
	RequestIdField      = "requestId"
	FunctionPrefixField = "functionPrefix"
)

type LogConfig struct {
	Level    logrus.Level
	Location io.Writer
}

var globalConfig LogConfig

func Initialize(lvlStr string, out io.Writer) (*LogConfig, error) {
	parsedLvl, err := ParseLevelFromStr(lvlStr)
	if err != nil {
		return nil, err
	}

	globalConfig = LogConfig{
		Level:    parsedLvl,
		Location: out,
	}
	return &globalConfig, nil
}

// CreateLogger returns a logrus FieldLogger instance.
// standardFlds is expected to be a Map At Least 1 entry (for :
// 1. the requestId (RequestIdField)
// 2. the function prefix (FunctionPrefixField)
// optionally, it can also have additional fields in form of key[string] = val[string]
func CreateLogger(fields map[string]string) (logrus.FieldLogger, error) {
	//check that at least "prefix" field was part of fields param
	fldCount := len(fields)
	if fldCount < 1 || fields[FunctionPrefixField] == "" {
		var err error
		errMsg := "error: Standard fields must contain at least 1 entry: '%s'"
		err = fmt.Errorf(errMsg, FunctionPrefixField)
		return nil, err
	}

	//Set Standard Fields - 2 Fields: requestId & functionPrefix
	var logrusLog = logrus.New()
	logrusLog.SetLevel(globalConfig.Level)
	logrusLog.SetOutput(globalConfig.Location)

	var logFields = map[string]interface{}{}
	logFields = logrus.Fields{
		RequestIdField:      fields[RequestIdField],
		FunctionPrefixField: fields[FunctionPrefixField],
	}

	// add additional fields, if any
	if fldCount > 2 {
		for k, v := range fields {
			if k != RequestIdField && k != FunctionPrefixField {
				logFields[k] = v
			}
		}
	}

	entry := logrusLog.WithFields(logFields)

	return entry, nil
}

func ParseLevelFromStr(lvlStr string) (logrus.Level, error) {
	level, err := logrus.ParseLevel(lvlStr)
	if err != nil {
		wrappedErr := fmt.Errorf("error parsing log Level - %w", err)
		return logrus.PanicLevel, wrappedErr ///in logrus, Default Level => Panic
	}
	return level, nil
}

// GetMyLogger is used by each Model/Function to generate it's own logger instance.
// NOTE: that if there is an error Creating the logger, we panic
func GetMyLogger(requestId string, prefix string) logrus.FieldLogger {
	if len(prefix) == 0 {
		errMsg := "ERROR: Could NOT acquire Logger: prefix value required"
		panic(errMsg)
	}

	stdFlds := map[string]string{
		RequestIdField:      requestId,
		FunctionPrefixField: prefix,
	}
	logger, err := CreateLogger(stdFlds)
	if err != nil {
		msg := "From " + prefix + ": ERROR: Could NOT acquire Logger: " + err.Error()
		fmt.Fprintf(os.Stderr, msg)
		panic(msg)
	}

	return logger
}
