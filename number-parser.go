package numberparser

import (
	"cmp"
	_ "embed"
	"log"

	"time"

	"slices"
	"strings"

	"github.com/gocarina/gocsv"
)

// Load the CSV and embed it into the binary so we do not have to worry about copying the
// csv file at package/distribution time.
// The embed directive loads the file into the binary and the next line makes it available
// into the variable `PhoneNumberDataCsv`
//
//go:embed prefix_data.csv
var PhoneNumberDataCsv string

var (
	PhoneNumberData    []PhoneNumberItem            = nil
	PhoneNumberDataMap map[string][]PhoneNumberItem = nil
)

// Represents each entry in the prefix_data.csv file
type PhoneNumberItem struct {
	RegionCode   string `csv:"region_code"`
	NumberPrefix string `csv:"number_prefix"`
	IsGeographic bool   `csv:"is_geographic"`
	IsMobile     bool   `csv:"is_mobile"`
	IsSatellite  bool   `csv:"is_satellite"`
}

// Loads the file prefix_data.csv into memory to allow number parsing via FindNumberDataForE164
func init() {
	var prefixKeys = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}

	// Clear any existing data

	PhoneNumberData = make([]PhoneNumberItem, 128000)
	PhoneNumberDataMap = make(map[string][]PhoneNumberItem, 9)

	for _, key := range prefixKeys {
		PhoneNumberDataMap[key] = make([]PhoneNumberItem, 20000)
	}

	ttx := time.Now()
	gocsv.UnmarshalStringToCallback(PhoneNumberDataCsv, func(item PhoneNumberItem) {
		for _, key := range prefixKeys {
			if strings.HasPrefix(item.NumberPrefix, key) {
				//log.Printf("Storing in i:%d key:`%s` <- %v", i, key, item)
				PhoneNumberDataMap[key] = append(PhoneNumberDataMap[key], item)
			}
		}
	})
	log.Printf("Finished Unpacking CSV into map size:%v ttx:%v", len(PhoneNumberDataMap), time.Since(ttx))

	ttx = time.Now()
	err := gocsv.UnmarshalString(PhoneNumberDataCsv, &PhoneNumberData)
	if err != nil {
		log.Fatalf("Unable to load the csv embed")
		return
	}
	log.Printf("Finished Unpacking CSV into array ttx:%v", time.Since(ttx))

	// Next, we should sort the data as follows:
	// Country then the largest prefix in descending order
	ttx = time.Now()
	slices.SortFunc(PhoneNumberData, func(i, j PhoneNumberItem) int {
		return cmp.Or(
			cmp.Compare(j.RegionCode, i.RegionCode),
			cmp.Compare(len(j.NumberPrefix), len(i.NumberPrefix)),
		)
	})
	log.Printf("Finished Sorting array ttx:%v", time.Since(ttx))

	// Sort the map
	ttx = time.Now()
	for _, key := range prefixKeys {
		slices.SortFunc(PhoneNumberDataMap[key], func(i, j PhoneNumberItem) int {
			return cmp.Or(
				cmp.Compare(i.RegionCode, j.RegionCode),               // ascending
				cmp.Compare(len(j.NumberPrefix), len(i.NumberPrefix)), // descending
			)
		})
	}
	log.Printf("Finished Sorting map ttx:%v", time.Since(ttx))
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

// Given a e164 argument, check if we have a match against our PhoneNumberData and return the PhoneNumberItem record.
// This function will normalize the argument to remove preceeding `+` or `0`
func FindNumberDataForE164v0(e164 string) *PhoneNumberItem {
	ttx := time.Now()

	e164 = SanitizeNumber(e164)

	i := slices.IndexFunc(PhoneNumberData, func(pnd PhoneNumberItem) bool {
		// For instance s -> `1` and our target is args.DestinationDdi
		// We want to check that we have s as a prefix of args.DestinationDdi
		//log.Printf(" - Examine prefix %s in e164:%s", pnd.NumberPrefix, e164)
		return len(pnd.NumberPrefix) > 0 && strings.HasPrefix(e164, pnd.NumberPrefix)
	})

	// The goal is to find the largest matching substring in the PhoneNumberData with the e164 argument
	// So for instance we have 5492314403 an entry that corresponds to Armenia mobile. Quite a large number
	// to perform a left-to-right match.
	// Start by matching the number exactly.

	if i != -1 {
		log.Printf("FindNumberDataForE164 - Found prefix @%v %v in e164:%s in ttx:%v", i, PhoneNumberData[i], e164, time.Since(ttx))
		return &PhoneNumberData[i]
	}

	log.Printf("FindNumberDataForE164 - Nothing Found prefix e164:%s in ttx:%v", e164, time.Since(ttx))

	return nil
}

// Given a e164 argument, check if we have a match against our PhoneNumberData and return the PhoneNumberItem record.
// This function will normalize the argument to remove preceeding `+` or `0`
// This is the improved search version using buckets to speed up lookup of number information.
// The previous version
func FindNumberDataForE164(e164 string) *PhoneNumberItem {
	ttx := time.Now()

	if e164 = SanitizeNumber(e164); len(e164) > 1 {
		firstPrefixCharacter := string([]rune(e164)[0])

		//log.Printf("FindNumberDataForE164v2 - e164:%v  `%v` in map len:%d of total:%d", e164, firstPrefixCharacter, len(PhoneNumberDataMap[firstPrefixCharacter]), len(PhoneNumberDataMap))

		i := slices.IndexFunc(PhoneNumberDataMap[firstPrefixCharacter], func(pnd PhoneNumberItem) bool {
			// For instance s -> `1` and our target is args.DestinationDdi
			// We want to check that we have s as a prefix of args.DestinationDdi
			//log.Printf("FindNumberDataForE164v2 - Examine prefix %s in e164:%s", pnd.NumberPrefix, e164)
			return len(pnd.NumberPrefix) > 0 && strings.HasPrefix(e164, pnd.NumberPrefix)
		})

		// The goal is to find the largest matching substring in the PhoneNumberData with the e164 argument
		// So for instance we have 5492314403 an entry that corresponds to Armenia mobile. Quite a large number
		// to perform a left-to-right match.
		// Start by matching the number exactly.

		if i != -1 {
			log.Printf("FindNumberDataForE164v2 - Found prefix @%d/%d -> %v in e164:%s in ttx:%v",
				i, len(PhoneNumberDataMap[firstPrefixCharacter]), PhoneNumberDataMap[firstPrefixCharacter][i],
				e164, time.Since(ttx))
			return &PhoneNumberDataMap[firstPrefixCharacter][i]
		}
	}
	log.Printf("FindNumberDataForE164v2 - Nothing Found prefix e164:%s in ttx:%v", e164, time.Since(ttx))

	return nil
}
