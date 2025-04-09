package numberparser

import (
	_ "embed"
	"strings"

	"github.com/gocarina/gocsv"
)

//	Load the CSV and embed it into the binary so we do not have to worry about copying the
//	csv file at package/distribution time.

//go:embed number-prefix-data.csv
var PhoneNumberDataCsv string

var (
	// Global instance 
	PhoneNumberCodex map[int]*CodexCountryItem = nil
)

// We use the CodexCountryItem structure as a means to shard the PhoneNumberItem objects.
type CodexCountryItem struct {
	ZoneId         int
	CountryCode    int
	LenCountryCode int
	MaxLenPrefix   int
	PrefixMap      map[string]*PhoneNumberItem
}

// Represents each entry in the prefix_data.csv file
type PhoneNumberItem struct {
	ZoneId          int    `csv:"zone_id"`
	CountryCode     int    `csv:"country_code"`
	RegionCode      string `csv:"region_code"`
	NumberPrefix    string `csv:"number_prefix"`
	IsGeographic    bool   `csv:"is_geographic"`
	IsMobile        bool   `csv:"is_mobile"`
	IsSatellite     bool   `csv:"is_satellite"`
	LenCountryCode  int    `csv:"cc_len"`
	LenNumberPrefix int    `csv:"prefix_len"`
}

// Helper function for PhoneNumberItem to get the Zone name based on the ZoneId member of the instance.
func (pni PhoneNumberItem) getZoneName() string {
	// Mapping uses the following source: https://en.wikipedia.org/wiki/List_of_telephone_country_codes#World_numbering_zones.
	var zoneNameTable = [...]string{
		// 0        1                                2         3         4         5                            6                             7                        8                      9
		"Unknown", "North American Numbering Plan", "Africa", "Europe", "Europe", "Central and South America", "Southeast Asia and Oceania", "Russia and Kazakhstan", "East and South Asia", "Middle East, Asia, Eastern Europe"}

	return zoneNameTable[pni.ZoneId]
}

// Helper function for the PhoneNumberItem to get the Zone group based on the ZoneId member of the instance.
func (pni PhoneNumberItem) getZoneGroup() string {
	// Colloquial grouping of the Zones for telecom use.
	var zoneGroupTable = [...]string{
		// 0        1           2      3         4         5           6       7         8       9
		"Unknown", "Americas", "MEA", "Europe", "Europe", "Americas", "APAC", "Russia", "APAC", "MEA"}

	return zoneGroupTable[pni.ZoneId]
}

// Loads the file prefix_data.csv into memory to allow number parsing via FindNumberDataForE164
func init() {
	// Clear any existing data
	PhoneNumberCodex = make(map[int]*CodexCountryItem)

	counterLines := 0
	counterCountries := 0

	gocsv.UnmarshalStringToCallback(PhoneNumberDataCsv, func(item PhoneNumberItem) {
		counterLines++

		// Initialize the map on-demand..
		cci := PhoneNumberCodex[item.CountryCode]
		if cci == nil {
			// Set the CodexCountryItem properties
			cci = &CodexCountryItem{ZoneId: item.ZoneId, CountryCode: item.CountryCode, LenCountryCode: item.LenCountryCode, PrefixMap: make(map[string]*PhoneNumberItem)}
			counterCountries++

			// Store the item
			PhoneNumberCodex[item.CountryCode] = cci
		}

		if cci != nil {
			// Keep track of the max prefix length for this country code
			if cci.MaxLenPrefix < item.LenNumberPrefix {
				cci.MaxLenPrefix = item.LenNumberPrefix
			}
			// Store the item
			cci.PrefixMap[item.NumberPrefix] = &item
		}
	})
	//log.Printf("Finished Unpacking CSV into codex size:%v ttx:%v", len(PhoneNumberCodex), time.Since(ttx))
	//log.Printf("Lines:%d  Zones:%d  Countries:%d  Prefixes:%d", counterLines, counterZones, counterCountries, counterPrefixes)
}

// Noramlizes the argument if it does not have a preceeding `+` then the result will have the `+` prefixed.
// If the argument already has `+` we return the argument as-is. This is *not* number validation!
func NormalizeE164(phone string) string {
	// Allocate the result rune slice and just copy items into it
	var processedPhone strings.Builder
	// Track the length of the result string
	var p int = 0
	// Track the previous character
	var previous rune

	// Make sure we have a leading `+`
	processedPhone.WriteRune('+')
	p++

	for i, v := range phone {
		switch v {
		case '0': // special case of removing 0 from the first digit
			if i != 0 && previous != '(' {
				processedPhone.WriteRune(v)
				p++
			}
			previous = v
		case '+', ' ', '-', '.', '(', ')', '/':
			previous = v // skip these items from result
		default:
			processedPhone.WriteRune(v)
			p++ // this is the length of the resulting string
			previous = v
		}
	}

	return processedPhone.String()
}

func SanitizeNumber(phone string) string {
	// Allocate the result rune slice and just copy items into it
	var processedPhone strings.Builder
	// Track the length of the result string
	var p int = 0
	// Track the previous character
	var previous rune

	for i, v := range phone {
		switch v {
		case '0': // special case of removing 0 from the first digit
			if i != 0 && previous != '(' {
				processedPhone.WriteRune(v)
				p++
			}
			previous = v
		case '+', ' ', '-', '(', ')', '/':
			previous = v // skip these items from result
		default:
			processedPhone.WriteRune(v)
			p++ // this is the length of the resulting string
			previous = v
		}
	}

	return processedPhone.String()
}

// This helper returns the 1 digit, two digit and three digit prefix from the given phone number.
// The country codes are integer values.
// The first return value is 1d, followed by 2d and last 3d
func getPossibleCountryCodes(str string) (int, int, int) {
	var cc1d int = 0
	var cc2d int = 0
	var cc3d int = 0

	for i, r := range str {
		if i == 0 {
			cc1d = int(r - '0')
			cc2d = int(r-'0') * 10
			cc3d = int(r-'0') * 100
		} else if i == 1 {
			cc2d = cc2d + int(r-'0')
			cc3d = cc3d + 10*int(r-'0')
		} else if i == 2 {
			cc3d = cc3d + int(r-'0')
		} else {
			// Bail out; we've gotten beyond the country code limit
			return cc1d, cc2d, cc3d
		}
	}

	return cc1d, cc2d, cc3d
}

// Returns the country code information matching the country code by scanning the first 1-3 digits of the argument
// It is required that you pass in result of
func FindCodexCountryItem(e164 string) *CodexCountryItem {
	cc1d, cc2d, cc3d := getPossibleCountryCodes(e164)

	if cci := PhoneNumberCodex[cc3d]; cci != nil {
		// First search for 3-digit country first
		return cci
	} else if cci := PhoneNumberCodex[cc2d]; cci != nil {
		// next search for 2-digit country
		return cci
	} else if cci := PhoneNumberCodex[cc1d]; cci != nil {
		// last search for 1-digit county codes
		return cci
	}

	return nil
}

// Given a e164 argument, check if we have a match against our PhoneNumberData and return the PhoneNumberItem record.
// This function will normalize the argument to remove preceeding `+` or `0`
// This is the improved search version using buckets to speed up lookup of number information.
func FindNumberDataForE164(e164 string) *PhoneNumberItem {
	var pni *PhoneNumberItem = nil

	if e164 = SanitizeNumber(e164); len(e164) > 1 {
		if cci := FindCodexCountryItem(e164); cci != nil {
			// Build the list of the prefixes that we should search with decreasing lengths
			for pfl := cci.MaxLenPrefix; pfl >= cci.LenCountryCode; pfl-- {
				if pfl > len(e164) {
					pfl = len(e164)
				}
				// Search for the given prefix
				if pni = cci.PrefixMap[e164[:pfl]]; pni != nil {
					return pni
				}
			}
		}
	}

	// this will return a nil
	return nil
}
