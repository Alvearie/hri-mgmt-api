package logwrapper

import (
	"bytes"
	//"bytes"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	TestRequestIdValue  = "xx6749493cj0"
	TestFuncPrefixValue = "module/fakeclass/fakefunction"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name        string
		lvlStr      string
		expectedLvl logrus.Level
		expectedLoc io.Writer
		expectedErr string
	}{
		{
			name:        "happy-path-debug-Level",
			lvlStr:      "debUG",
			expectedLvl: logrus.DebugLevel,
			expectedLoc: os.Stdout,
		},
		{
			name:        "warn-Level",
			lvlStr:      "Warn",
			expectedLvl: logrus.WarnLevel,
			expectedLoc: os.Stdout,
		},
		{
			name:        "Invalid-log-Level",
			lvlStr:      "Invalid-Log-Lev8l1289", //we do NOT send in a valid Level
			expectedErr: "error parsing log Level - not a valid logrus Level: \"Invalid-Log-Lev8l1289\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logger, err = Initialize(tt.lvlStr, os.Stdout)
			if err != nil {
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.Equal(t, tt.expectedLvl, logger.Level,
					"Level '%v' is Not equal to Level '%v'", logger.Level.String(), tt.expectedLvl)
			}
		})
	}
}

func TestCreateLogger(t *testing.T) {
	var missingRequestId = make(map[string]string)
	missingRequestId[FunctionPrefixField] = "some/functionName"

	var additionalFields = getDefaultFields()
	additionalFields["addlField"] = "monkeeZ"

	var missingPrefixFlds = make(map[string]string)
	missingPrefixFlds[RequestIdField] = TestRequestIdValue
	missingPrefixFlds["addlField"] = "testMonkee1"

	tests := []struct {
		name            string
		fields          map[string]string
		expectedErr     string
		expectedLogMsg  string
		extraFieldMatch string
	}{
		{
			name:           "success-path",
			fields:         getDefaultFields(),
			expectedLogMsg: "level=info msg=\"Success CreateLogger\" functionPrefix=module/fakeclass/fakefunction requestId=xx6749493cj0\n",
		},
		{
			name:        "error-no-fields",
			fields:      make(map[string]string),
			expectedErr: "error: Standard fields must contain at least 1 entry: 'functionPrefix'",
		},
		{
			name:        "error-no-prefix-field",
			fields:      missingPrefixFlds,
			expectedErr: "error: Standard fields must contain at least 1 entry: 'functionPrefix'",
		},
		{
			name:           "success-no-requestId",
			fields:         missingRequestId,
			expectedLogMsg: "level=info msg=\"Success CreateLogger\" functionPrefix=some/functionName requestId=\n",
		},
		{
			name:            "success-additional-non-standard-fields",
			fields:          additionalFields,
			expectedLogMsg:  "functionPrefix=module/fakeclass/fakefunction requestId=xx6749493cj",
			extraFieldMatch: "level=info msg=\"Success CreateLogger\" addlField=monkeeZ",
		},
	}

	var buf bytes.Buffer
	Initialize("info", &buf)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logger, err = CreateLogger(tt.fields)
			if len(tt.expectedErr) > 0 {
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				//Match string from buffer contains expectedLog Msg (substring)
				logger.Info("Success CreateLogger")
				output := buf.String()
				assert.Contains(t, output, tt.expectedLogMsg)
				if len(tt.fields) > 2 {
					assert.Contains(t, output, tt.extraFieldMatch)
				}
			}
		})
		buf.Reset()
	}
}

func TestGetMyLogger(t *testing.T) {
	var oneMissingField = make(map[string]string)
	oneMissingField[FunctionPrefixField] = "some/functionName"
	testReqId := "12833eQwP20"
	testFuncPrefix := "some/other/fakefunction"
	emptyStr := ""
	successMsgStartStr := "level=info msg=\"Success calling GetMyLogger\" functionPrefix="
	defaultLogMsg := successMsgStartStr + testFuncPrefix + " requestId=" + testReqId + "\n"

	tests := []struct {
		name           string
		requestId      string
		prefix         string
		expectedErr    string
		expectedLogMsg string
	}{
		{
			name:           "success-path",
			requestId:      testReqId,
			prefix:         testFuncPrefix,
			expectedLogMsg: defaultLogMsg,
		},
		{
			name:           "happy-path-empty-request-id",
			requestId:      emptyStr,
			prefix:         testFuncPrefix,
			expectedLogMsg: successMsgStartStr + testFuncPrefix + " requestId=\n",
		},
		{
			name:        "error-missing-func-prefix",
			requestId:   testReqId,
			prefix:      emptyStr,
			expectedErr: "ERROR: Could NOT acquire Logger: prefix value required",
		},
	}

	var buf bytes.Buffer
	Initialize("info", &buf)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expectedErr) > 0 {
				assert.PanicsWithValue(t, tt.expectedErr, func() { GetMyLogger(tt.requestId, tt.prefix) })
			} else {
				var logger = GetMyLogger(tt.requestId, tt.prefix)
				//Match string from buffer contains expectedLog Msg (substring)
				logger.Info("Success calling GetMyLogger")
				output := buf.String()
				assert.Contains(t, output, tt.expectedLogMsg)
				buf.Reset()
			}
		})
	}
}

func TestGetLoggerAndAddMoreFields(t *testing.T) {
	assertions := assert.New(t)

	var buf bytes.Buffer
	Initialize("info", &buf)
	var initLogger, _ = CreateLogger(getDefaultFields())

	key1 := "batchId"
	value1 := "catzBatch88"
	key2 := "tenantId"
	value2 := "tenZxC103"
	newLogger := initLogger.WithFields(logrus.Fields{
		key1: value1,
		key2: value2,
	})

	msg := "additional fields logger test"
	key1Match := key1 + "=" + value1
	key2Match := key2 + "=" + value2
	matchStr1 := "level=warning msg=\"" + msg + "\""
	matchStr2 := "functionPrefix=module/fakeclass/fakefunction requestId=xx6749493cj0"
	newLogger.Warn(msg)
	assertions.Contains(buf.String(), matchStr1)
	assertions.Contains(buf.String(), matchStr2)
	assertions.Contains(buf.String(), key1Match)
	assertions.Contains(buf.String(), key2Match)
}

func TestParseLevelFromStr(t *testing.T) {
	assertions := assert.New(t)
	var levelStr1 = "debug"
	var levelStr2 = "ErrOR"
	var levelStr3 = "Panic"

	lvl1, err := ParseLevelFromStr(levelStr1)
	if err != nil {
		assertions.Fail("ParseLevelFromStr failed for %v; Error: %v", levelStr1, err.Error())
	}
	assertions.Equal(logrus.DebugLevel, lvl1)

	lvl2, err := ParseLevelFromStr(levelStr2)
	if err != nil {
		assertions.Fail("ParseLevelFromStr failed for %v; Error: %v", levelStr2, err.Error())
	}
	assertions.Equal(logrus.ErrorLevel.String(), lvl2.String())

	lvl3, err := ParseLevelFromStr(levelStr3)
	if err != nil {
		assertions.Fail("ParseLevelFromStr failed for %v; Error: %v", levelStr3, err.Error())
	}
	assertions.Equal(logrus.PanicLevel.String(), lvl3.String())

	//Force Return error
	var invalidLevel = "InVal1D-Level"
	_, err = ParseLevelFromStr(invalidLevel)
	if err == nil {
		assertions.Fail("Expected Error ParseLevelFromStr failed for %v; Error: %v", invalidLevel)
	}
	assertions.Equal("error parsing log Level - not a valid logrus Level: \"InVal1D-Level\"", err.Error())
}

func getDefaultFields() map[string]string {
	var fields = make(map[string]string)
	fields[RequestIdField] = TestRequestIdValue
	fields[FunctionPrefixField] = TestFuncPrefixValue
	return fields
}
