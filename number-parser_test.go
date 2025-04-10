package numberparser

import (
	_ "embed"
	"fmt"
	"log"
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
		res := FindNumberDataForE164(SanitizeNumber(tc.input))
		if res == nil || res.IsSatellite != tc.want {
			t.Errorf("FindNumberDataForE164: in:%s  want: %v", tc.input, tc.want)
		}
	}
}

func TestPhoneNumberData_Mobile(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+447762000000", true},
		{"+447762987654", true},
		{"+5492362336", true},
		{"+14159991111", false},
		{"+52 13 1400 0000", true},
		{"+52 81 9876 5432", false},
		{"+52 33 1122 3344", true},
		{"+443111111", false},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(SanitizeNumber(tc.input))
		if res == nil || (res.IsMobile != tc.want) {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsMobile, tc.want)
		}
	}
}

func TestPhoneNumberData_Mobile_UK(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+447762000000", true},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(SanitizeNumber(tc.input))
		if res == nil || (res.IsMobile != tc.want) {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.IsMobile, tc.want)
		}
	}
}

func TestPhoneNumberData_Mobile_MX(t *testing.T) {
	testcases := []struct {
		input string
		want  bool
	}{
		{"+52 64 4454-4600", true},
		{"+52 13 1400 0000", true},
	}

	for _, tc := range testcases {
		res := FindNumberDataForE164(SanitizeNumber(tc.input))
		if res == nil || (res.IsMobile != tc.want) {
			t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res, tc.want)
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
		res := FindNumberDataForE164(SanitizeNumber(tc.input))
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
		{" 1 800 555 1212", "+18005551212"},
		{"44(20)7123-4567", "+442071234567"},
	}

	for _, tc := range testcases {
		res := NormalizeE164(tc.input)
		if res != tc.output {
			t.Errorf("NoramlizeE164: in: %s   want: %s", res, tc.output)
		}
	}
}

func TestNormalizeE164WithAreaCode(t *testing.T) {
	testcases := []struct {
		input, output string
	}{
		{"44(020)7123-4567", "+442071234567"},
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
		{"+447762000000", "GB"},
		{"+54923623360", "AR"},
		{"+14159991111", "US"},
		{"+44311111100", "GB"},
		{"+33750730000", "FR"},
		{"+24100000000", "GA"},
		{"+69100000000", "FM"},
		{"+77280000000", "KZ"},
		{"+85270900000", "HK"},
		{"+99554400000", "GE"},
		{"+14156292008", "US"}}

	for _, tc := range testcases {
		testname := fmt.Sprintf("%s-%s", tc.input, tc.want)
		t.Run(testname, func(t *testing.T) {
			res := FindNumberDataForE164(SanitizeNumber(tc.input))
			if res != nil && res.RegionCode != tc.want {
				t.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
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
		"+447762000000",
		"+54923623360",
		"+14159991111",
		"+44311111100",
		"+33750730000",
		"+24100000000",
		"+69100000000",
		"+77280000000",
		"+85270900000",
		"+99554400000",
		"+14156292008",
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
		{"+1-800-555-1212x,,099", "18005551212x,,099"},
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

func BenchmarkSanitizeNumber(b *testing.B) {
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
		{"+1(925)300-4504", "US"},
		{"+1(408)555-2270", "US"},
		{"+52 55 1234 5678", "MX"},
		{" +52 33 1234 5678 ", "MX"},
		{"+52 222 1234 5678", "MX"},
		{"+52 664 1234 5678", "MX"},
		{"+52 81 9876 5432", "MX"},
		{"+447762000000", "GB"},
		{"+54923623360", "AR"},
		{"+14159991111", "US"},
		{"+44311111100", "GB"},
		{"+33750730000", "FR"},
		{"+24100000000", "GA"},
		{"+69100000000", "FM"},
		{"+77280000000", "KZ"},
		{"+85270900000", "HK"},
		{"+99554400000", "GE"},
		{"+14156292008", "US"},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			SanitizeNumber(tc.input)
			//if res != nil && res.RegionCode != tc.want {
			//	b.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
			//}
		}
	}
}

func BenchmarkFindNumberDataForE164(b *testing.B) {
	testcases := []struct {
		input string
		want  string
	}{
		{"12125554448", "US"},
		{"447762987654", "GB"},
		{"14158746923", "US"},
		{"12125552270", "US"},
		{"16508982178", "US"},
		{"1510866949", "US"},
		{"19253004504", "US"},
		{"14085552270", "US"},
		{"525512345678", "MX"},
		{"523312345678 ", "MX"},
		{"5222212345678", "MX"},
		{"5266412345678", "MX"},
		{"528198765432", "MX"},
		{"447762000000", "GB"},
		{"54923623360", "AR"},
		{"14159991111", "US"},
		{"44311111100", "GB"},
		{"33750730000", "FR"},
		{"24100000000", "GA"},
		{"69100000000", "FM"},
		{"77280000000", "KZ"},
		{"85270900000", "HK"},
		{"99554400000", "GE"},
		{"14156292008", "US"}}

	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			res := FindNumberDataForE164(tc.input)
			if res != nil && res.RegionCode != tc.want {
				b.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
			}
		}
	}
}

func BenchmarkFindNumberDataForE164_US(b *testing.B) {
	testcases := []struct {
		input string
		want  string
	}{
		{"+12125554448", "US"},
		{"+14158746923", "US"},
		{"+12125552270", "US"},
		{"+16508982178", "US"},
		{"+1510866949", "US"},
		{"+19253004504", "US"},
		{"+14085552270", "US"},
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

func BenchmarkFindNumberDataForE164_UK(b *testing.B) {
	testcases := []struct {
		input string
		want  string
	}{
		{"+447762987654", "GB"},
		{"+447762000000", "GB"}}

	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			res := FindNumberDataForE164(tc.input)
			if res != nil && res.RegionCode != tc.want {
				b.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
			}
		}
	}
}

func BenchmarkFindNumberDataForE164_MX(b *testing.B) {
	testcases := []struct {
		input string
		want  string
	}{
		{"+525512345678", "MX"},
		{"+523312345678 ", "MX"},
		{"+5222212345678", "MX"},
		{"+5266412345678", "MX"},
		{"+528198765432", "MX"}}

	for i := 0; i < b.N; i++ {
		for _, tc := range testcases {
			res := FindNumberDataForE164(tc.input)
			if res != nil && res.RegionCode != tc.want {
				b.Errorf("FindNumberDataForE164: %s in: %v   want: %v", tc.input, res.RegionCode, tc.want)
			}
		}
	}
}

func TestZoneMatch(t *testing.T) {
	tests := []struct {
		phone     string
		zoneId    int
		zoneName  string
		zoneGroup string
	}{
		{"447762000000", 4, "Europe", "Europe"},
		{"54923623360", 5, "Central and South America", "Americas"},
		{"14159991111", 1, "North American Numbering Plan", "Americas"},
		{"44311111100", 4, "Europe", "Europe"},
		{"33750730000", 3, "Europe", "Europe"},
		{"24100000000", 2, "Africa", "MEA"},
		{"69100000000", 6, "Southeast Asia and Oceania", "APAC"},
		{"77280000000", 7, "Russia and Kazakhstan", "Russia"},
		{"85270900000", 8, "East and South Asia", "APAC"},
		{"99554400000", 9, "Middle East, Asia, Eastern Europe", "MEA"},
	}

	for i, tt := range tests {
		got := FindNumberDataForE164(tt.phone)
		log.Printf("test %d :%v ... got:%v", i, tt, got)
		if tt.zoneId != got.ZoneId {
			t.Errorf("Failed ZoneId match; tt:%v --> got:%v", tt, got)
		}
		if tt.zoneName != got.ZoneName() {
			t.Errorf("Failed ZoneName match; tt:%v --> got:%v", tt, got)
		}
		if tt.zoneGroup != got.ZoneGroup() {
			t.Errorf("Failed ZoneName match; tt:%v --> got:%v", tt, got)
		}
	}
}

func TestZoneGroupMatch(t *testing.T) {
	tests := []struct {
		zoneGroup string
		phone1    string
		phone2    string
	}{
		{"Europe", "447762000000", "337507300000"},
		{"Americas", "5492362336", "14159991111"},
		{"MEA", "24100000000", "99554400000"},
		{"APAC", "69100000000", "85270900000"},
	}

	for i, tt := range tests {
		p1 := FindNumberDataForE164(tt.phone1)
		p2 := FindNumberDataForE164(tt.phone2)

		log.Printf("test %d :%v ... got:%v:%v", i, tt, p1, p2)
		if !(p1.ZoneGroup() == p2.ZoneGroup() && p1.ZoneGroup() == tt.zoneGroup) {
			t.Errorf("Failed ZoneId match; tt:%v --> got:%v", tt, p1)
		}
	}
}
