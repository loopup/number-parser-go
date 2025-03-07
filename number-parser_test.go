package numberparser

import (
	_ "embed"
	
	
	"strings"
	"testing"
)

func TestPhoneNumberDataCsv_NotEmpty(t *testing.T) {
	if len(PhoneNumberDataCsv) == 0 {
		t.Errorf("PhoneNumberDataCsv is empty!")
	}
}

func TestPhoneNumberData_Parsed(t *testing.T) {
	if len(PhoneNumberData) == 0 {
		t.Errorf("PhoneNumberData is empty!")
	}
}

func TestPhoneNumberData_Invalid(t *testing.T) {
	res := FindNumberDataForE164("+21012345")
	if res != nil {
		t.Errorf("PhoneNumberData must be empty for invalid number!")
	}
}

func TestPhoneNumberData_UKMobile(t *testing.T) {
	res := FindNumberDataForE164("+447762000000")
	if res != nil && !res.IsMobile {
		t.Errorf("PhoneNumberData must be mobile! %v", res)
	}
}

func TestPhoneNumberData_UKMobile_old(t *testing.T) {
	res := FindNumberDataForE164v0("+447762987654")
	if res != nil && !res.IsMobile {
		t.Errorf("PhoneNumberData must be mobile! %v", res)
	}
}

func TestPhoneNumberData_Satellite(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+87810", true},
		{"+614899", true},
		{"+4312345", false},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(tc.input)
		if res.IsSatellite != tc.want {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsSatellite, tc.want)
		}
	}
}

func TestPhoneNumberData_Mobile(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+447762000000", true},
		{"+5492362336", true},
		{"+14159991111", false},
		{"+443111111", false},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(tc.input)
		if res.IsMobile != tc.want {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsMobile, tc.want)
		}
	}
}

func TestPhoneNumberData_Geographic(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+441387", true},
		{"+4420000", true},
		{"+1415", true},
		{"+13102223333", true},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(tc.input)
		if res.IsGeographic != tc.want {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsGeographic, tc.want)
		}
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

func TestSanitizeNumber(t *testing.T) {
	tests := []struct {
		phone string
		want  string
	}{
		{"+18005551212", "18005551212"},
		{"+18005551212x,,099", "18005551212x,,099"},
		{"1 (800) 555-1212", "18005551212"},
		{" 1 (800) 555-1212", "18005551212"},
		{"018005551212", "18005551212"},
	}

	for _, tt := range tests {
		got := SanitizeNumber(tt.phone)
		if got != tt.want {
			t.Errorf("SanitizeNumber() = %v, want %v", got, tt.want)
		}
		t.Logf("SanitizeNumber(%v) = %v, want %v", tt.phone, got, tt.want)
	}
}

func FuzzSanitizeNumber(f *testing.F) {
	testcases := []string{"18005551212",
		"+18005551212",
		" 18005551212",
		"08005551212",
		"+18005551212x,,099",
		"1 (800) 555-1212",
		" 1 (800) 555-1212",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, arg string) {
		res := SanitizeNumber(arg)
		if len(res) == 0 {
			t.Errorf("NormalizeE164 fails %s --> %s", arg, res)
		}
	})
}
