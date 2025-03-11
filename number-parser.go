package numberparser

import (
	_ "embed"

	"strings"

	"github.com/gocarina/gocsv"
)

// Load the CSV and embed it into the binary so we do not have to worry about copying the
// csv file at package/distribution time.
// The embed directive loads the file into the binary and the next line makes it available
// into the variable `PhoneNumberDataCsv`
//
//go:embed number-prefix-data.csv
var PhoneNumberDataCsv string

var (
	PhoneNumberCodex map[string]*CodexCountryItem = nil // zone :-> country :-> array PhoneNumberItem
)

type CodexCountryItem struct {
	ZoneId         rune
	CountryCode    string
	LenCountryCode int
	MaxLenPrefix   int
	PrefixMap      map[string]*PhoneNumberItem
}

// Represents each entry in the prefix_data.csv file
type PhoneNumberItem struct {
	ZoneId          rune   `csv:"zone_id"`
	CountryCode     string `csv:"country_code"`
	RegionCode      string `csv:"region_code"`
	NumberPrefix    string `csv:"number_prefix"`
	IsGeographic    bool   `csv:"is_geographic"`
	IsMobile        bool   `csv:"is_mobile"`
	IsSatellite     bool   `csv:"is_satellite"`
	LenCountryCode  int    `csv:"cc_len"`
	LenNumberPrefix int    `csv:"prefix_len"`
}

// Loads the file prefix_data.csv into memory to allow number parsing via FindNumberDataForE164
func init() {
	// Clear any existing data
	PhoneNumberCodex = make(map[string]*CodexCountryItem)

	counterLines := 0
	counterCountries := 0
	//counterZones := 0
	//counterPrefixes := 0

	//ttx := time.Now()
	gocsv.UnmarshalStringToCallback(PhoneNumberDataCsv, func(item PhoneNumberItem) {
		counterLines++
		//log.Printf("ZoneId: %v CountryCode: %v Prefix:%v <-- %v", item.ZoneId, item.CountryCode, item.NumberPrefix, item)
		// Initialize the map on-demand..

		cci := PhoneNumberCodex[item.CountryCode]
		if cci == nil {
			cci = &CodexCountryItem{CountryCode: item.CountryCode, LenCountryCode: item.LenCountryCode, ZoneId: item.ZoneId, PrefixMap: make(map[string]*PhoneNumberItem)}
			counterCountries++
			//log.Printf("  Stored new prefix. len:%d  item: %v", len(cci.PrefixMap), cci.PrefixMap[item.NumberPrefix])
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
	// Returns the number iff it has a ` ` prefix
	phone = SanitizeNumber(phone)
	return "+" + phone
}

// Remove all decorations from the number
func SanitizeNumber(phone string) string {
	// There is a special case where the leading 0 must be removed.
	phone, _ = strings.CutPrefix(phone, "0")
	// Remove all the items with their corresponding substitutions.
	// Leave the "x" and ","
	return strings.NewReplacer("-", "", // remove all dashes
		"+", "", // remove all +
		"(", "", // remove all left-paren
		")", "", // remove all right-paren
		" ", "").Replace(phone) // remove all spaces
}

// Returns the country code information matching the country code by scanning the first 1-3 digits of the argument
// It is required that you pass in result of
func FindCodexCountryItem(e164 string) *CodexCountryItem {
	cc1d := e164[:1]
	cc2d := e164[:2]
	cc3d := e164[:3]

	if cc1d != "+" {
		// Search for two-digit country first
		// next search for 1-digit country
		// last search for 3-digit county codes
		// This approach is based on the frequency of use and the
		// size of the dataset.
		if cci := PhoneNumberCodex[cc2d]; cci != nil {
			return cci
		} else if cci := PhoneNumberCodex[cc1d]; cci != nil {
			return cci
		} else if cci := PhoneNumberCodex[cc3d]; cci != nil {
			return cci
		}
	}

	return nil
}

// Given a e164 argument, check if we have a match against our PhoneNumberData and return the PhoneNumberItem record.
// This function will normalize the argument to remove preceeding `+` or `0`
// This is the improved search version using buckets to speed up lookup of number information.
// The previous version
func FindNumberDataForE164(e164 string) *PhoneNumberItem {
	if e164 = SanitizeNumber(e164); len(e164) > 1 {
		if cci := FindCodexCountryItem(e164); cci != nil {
			for pfl := cci.MaxLenPrefix; pfl >= cci.LenCountryCode; pfl-- {
				if pfl > len(e164) {
					pfl = len(e164)
				}
				pf := e164[:pfl]
				//log.Printf("  search e164:%s  pfl:%v  pf:%v  cclen:%v mapsize:%d..", e164, pfl, pf, cci.LenCountryCode, len(cci.PrefixMap))
				if pfm := cci.PrefixMap[pf]; pfm != nil {
					return pfm
				}
			}
		}
	}
	return nil
}
