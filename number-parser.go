package numberparser

import (
	_ "embed"
	"log"
	"slices"
	"strings"

	"github.com/gocarina/gocsv"
)

// Load the CSV and embed it into the binary so we do not have to worry about copying the
// csv file at package/distribution time.
// The embed directive loads the file into the binary and the next line makes it available
// into the variable `PhoneNumberDataCsv`
//go:embed prefix_data.csv
var PhoneNumberDataCsv string

var (
	PhoneNumberData []PhoneNumberItem
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
	// Clear any existing data
	PhoneNumberData = nil

	err := gocsv.UnmarshalString(PhoneNumberDataCsv, &PhoneNumberData)
	if err != nil {
		log.Fatalf("Unable to load the csv embed")
		return
	}
}

// Noramlizes the argument if it does not have a preceeding `+` then the result will have the `+` prefixed.
// If the argument already has `+` we return the argument as-is. This is *not* number validation!
func NormalizeE164(phone string) string {
	// Returns the number iff it has a ` ` prefix
	phone, _ = strings.CutPrefix(phone, " ")

	// If we do not have the prefix `+` then add it..
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	return phone
}

// Given a e164 argument, check if we have a match against our PhoneNumberData and return the PhoneNumberItem record.
// This function will normalize the argument to remove preceeding `+` or `0`
func FindNumberDataForE164(e164 string) PhoneNumberItem {
	e164, _ = strings.CutPrefix(e164, "+")
	e164, _ = strings.CutPrefix(e164, "0")

	i := slices.IndexFunc(PhoneNumberData, func(pnd PhoneNumberItem) bool {
		// For instance s -> `1` and our target is args.DestinationDdi
		// We want to check that we have s as a prefix of args.DestinationDdi
		return strings.HasPrefix(e164, pnd.NumberPrefix)
	})

	if i != -1 {
		return PhoneNumberData[i]
	}

	return PhoneNumberItem{}
}
