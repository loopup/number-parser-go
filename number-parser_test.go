package numberparser

import (
	"strings"
	"testing"
)

func TestPhoneNumberDataCsv_NotEmpty(t *testing.T){
	if len(PhoneNumberDataCsv) == 0  {
		t.Errorf("PhoneNumberDataCsv is empty!")
	}
}

func TestPhoneNumberData_Parsed(t *testing.T){
	if len(PhoneNumberData) == 0  {
		t.Errorf("PhoneNumberData is empty!")
	}
}


func TestNormalizeE164(t *testing.T) {
	testcases := []struct {
		input, output string
	}{
		{"18005551212", "+18005551212"},
		{" 18005551212", "+18005551212"},
		{"442071234567", "+442071234567"},
	}

	for _, tc := range testcases {
		res := NormalizeE164(tc.input)
		if res != tc.output {
			t.Errorf("NoramlizeE164: in: %s   want: %s", res, tc.output)
		}
	}
}

func FuzzNormalizeE164(f *testing.F) {
	testcases := []string{"18005551212", "+18005551212", " 18005551212", "08005551212"}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, arg string) {
		res := NormalizeE164(arg)
		if !strings.HasPrefix(res, "+") {
			t.Errorf("NormalizeE164 fails %s --> %s", arg, res)
		}
	})
}
