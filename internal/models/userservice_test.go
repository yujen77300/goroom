package models

import (
	"testing"
)

var emailTestCases = []struct {
	email  string
	result bool
}{
	{"example123@gmail.com", true},
	{"example@com", false},
	{"example-email", false},
	{"@example.com", false},
	{"example$@gmail.com", false},
	{"example.com", false},
	{"example+email@gmail.com", false},
}

func TestEmailValidation(t *testing.T) {
	for _, eachCase := range emailTestCases {
		eachResult := emailValidation(eachCase.email)
		if eachResult != eachCase.result {
			t.Errorf("Expected %v, but got %v for email %s", eachCase.result, eachResult, eachCase.email)
		}
	}
}

var pwdTestCases = []struct {
	pwd    string
	result bool
}{
	{"abcd1234", true},
	{"A1b2c3d4", true},
	{"abcdefg", false},
	{"12345678", false},
	{"abc123", false},
	{"", false},
	{"@12345678", false},
	{"@ab13245678", true},
}

func TestPwdValidation(t *testing.T) {
	for _, eachCase := range pwdTestCases {
		eachResult := pwdValidation(eachCase.pwd)
		if eachResult != eachCase.result {
			t.Errorf("Expected %v, but got %v for password %s", eachCase.result, eachResult, eachCase.pwd)
		}
	}
}
