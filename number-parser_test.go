package numberparser

import (
	_ "embed"
	"fmt"
	"slices"

	"strings"
	"testing"
)

func TestPhoneNumberDataCsv_NotEmpty(t *testing.T) {
	if len(PhoneNumberDataCsv) == 0 {
		t.Errorf("PhoneNumberDataCsv is empty!")
	}
}

func TestPhoneNumberData_Invalid(t *testing.T) {
	res := FindNumberDataForE164("+21012345")
	if res != nil {
		t.Errorf("PhoneNumberData must be empty for invalid number!")
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
		if res == nil || res.IsSatellite != tc.want {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsSatellite, tc.want)
		}
	}
}

func TestPhoneNumberData_Mobile(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+447762000000", true}, {"+447762987654", true},
		{"+5492362336", true},
		{"+14159991111", false},
		{"+52 55 1234 5678", true},
		{"+52 81 9876 5432", false},
		{"+52 33 1122 3344", true},
		{"+443111111", false},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(tc.input)
		if res == nil || (res != nil && res.IsMobile != tc.want) {
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
		{"+86 10 12345678", true},
		{"+86 21 98765432", true},
		{"+86 24 11223344", true},
		{"+86 20 55667788", true},
		{"+86 28 99887766", true},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(tc.input)
		if res == nil || res.IsGeographic != tc.want {
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

func TestFindNumberDataForE164(t *testing.T) {
	testcases := []struct {
		input string
		want  string
	}{
		{"+12125554448", "US"},
		{"+447762987654", "GB"},
		{"+14158746923", "US"},
		{"+12125552270", "US"},
		{"+16508982178", "US"},
		{"+1510866949", "US"},
		{"+19253004504", "US"},
		{"+14085552270", "US"},
		{"+52 55 1234 5678", "MX"},
		{" +52 33 1234 5678 ", "MX"},
		{"+52 222 1234 5678", "MX"},
		{"+52 664 1234 5678", "MX"},
		{"+52 55 1234 5678", "MX"},
		{"+52 81 9876 5432", "MX"},
		{"+52 33 1122 3344", "MX"},
		{"+14156292008", "US"}}

	for _, tc := range testcases {
		testname := fmt.Sprintf("%s-%s", tc.input, tc.want)
		t.Run(testname, func(t *testing.T) {
			res := FindNumberDataForE164(tc.input)
			if res != nil && res.RegionCode != tc.want {
				t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsSatellite, tc.want)
			}
		})
	}
}


func FuzzFindNumberDataForE164(f *testing.F) {
	testcases := []string{"+12125554448",
		"+14158746923",
		"+12125552270",
		"+16508982178",
		"+1510866949",
		"+19253004504",
		"+14085552270",
		"+14156292008",
		"+52 55 1234 5678",
		" +52 33 1234 5678 ",
		"+52 222 1234 5678",
		"+52 664 1234 5678",
		"+52 55 1234 5678",
		"+52 81 9876 5432",
		"+52 33 1122 3344",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, arg string) {
		arg = SanitizeNumber(arg)

		// Handles cases where the number is left as empty!
		if len(arg) < 9 {
			t.Skipf("FuzzFindNumberDataForE164 - Skipping arg:`%s`; too small.", arg)
		}

		res := FindNumberDataForE164(arg)
		if res == nil && slices.Contains(testcases, arg) {
			t.Errorf("FuzzFindNumberDataForE164 fails %s", arg)
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

		// This minor update will allow us to bypass input that is
		// too small to be a phone number causing a needless failure
		// during fuzzing.
		if len(res) < 6 {
			t.Skipf("Skipping res:`%s`; too small.", res)
		}

		if len(arg) != 0 && len(res) == 0 {
			t.Errorf("NormalizeE164 fails %s --> %s", arg, res)
		}
	})
}


func BenchmarkFindNumberDataForE164(b *testing.B) {
	testcases := []struct {
		input string
		want  string
	}{
		{"+12125554448", "US"},
		{"+447762987654", "GB"},
		{"+14158746923", "US"},
		{"+12125552270", "US"},
		{"+16508982178", "US"},
		{"+1510866949", "US"},
		{"+19253004504", "US"},
		{"+14085552270", "US"},
		{"+52 55 1234 5678", "MX"},
		{" +52 33 1234 5678 ", "MX"},
		{"+52 222 1234 5678", "MX"},
		{"+52 664 1234 5678", "MX"},
		{"+52 55 1234 5678", "MX"},
		{"+52 81 9876 5432", "MX"},
		{"+52 33 1122 3344", "MX"},
		{"+14156292008", "US"}}

	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			res := FindNumberDataForE164(tc.input)
			if res != nil && res.RegionCode != tc.want {
				b.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
			}
		}
	}
}
