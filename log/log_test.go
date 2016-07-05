package log

import "testing"

func TestSetLevelByName(t *testing.T) {
	if err := SetLevelByName("INFO"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("DEBUG"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("WARN"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("ERROR"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("FATAL"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("INFo"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("DebUG"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("WaRN"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("ERrOR"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("fatAL"); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevelByName("fasstAL"); err == nil {
		t.Fail()
	}

	SetLevelByName("faTal")
	if logLevel != LevelFatal {
		t.Error("log level mismatch")
		t.Fail()
	}
}

func TestSetLevel(t *testing.T) {
	if err := SetLevel(LogLevel(1)); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevel(LogLevel(2)); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevel(LogLevel(3)); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevel(LogLevel(4)); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevel(LogLevel(5)); err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	if err := SetLevel(LogLevel(0)); err == nil {
		t.Error("invalid log level")
		t.Fail()
	}

	if err := SetLevel(LogLevel(6)); err == nil {
		t.Error("invalid log level")
		t.Fail()
	}

	if err := SetLevel(LogLevel(255)); err == nil {
		t.Error("invalid log level")
		t.Fail()
	}

	SetLevel(LogLevel(4))
	if logLevel != LevelError {
		t.Error("log level mismatch")
		t.Fail()
	}
}

func TestLogLevel_String(t *testing.T) {
	if LevelInfo.String() != "INFO" {
		t.Errorf("wrong level string: %s", LevelInfo.String())
		t.Fail()
	}

	if LevelDebug.String() != "DEBUG" {
		t.Errorf("wrong level string: %s", LevelInfo.String())
		t.Fail()
	}

	if LevelWarn.String() != "WARN" {
		t.Errorf("wrong level string: %s", LevelInfo.String())
		t.Fail()
	}

	if LevelError.String() != "ERROR" {
		t.Errorf("wrong level string: %s", LevelInfo.String())
		t.Fail()
	}

	if LevelFatal.String() != "FATAL" {
		t.Errorf("wrong level string: %s", LevelInfo.String())
		t.Fail()
	}
}
