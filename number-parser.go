package numberparsergo

import (
	"os"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

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
func LoadNumberParserCache() {
	ttx := time.Now()
	inf, err := os.Open("prefix_data.csv")
	if err != nil {
		slog.With("func", "LoadNumberParserCache").With("ttx", time.Since(ttx)).Error("Unable to load the `prefix_data.csv`", "err", err)
		return
	}

	defer inf.Close()

	// Clear any existing data
	PhoneNumberData = nil

	err = gocsv.UnmarshalFile(inf, &PhoneNumberData)
	if err != nil {
		slog.With("func", "LoadNumberParserCache").With("ttx", time.Since(ttx)).Error("Failed decoding the prefix_data.csv", "err", err)
		return
	}

	slog.With("func", "LoadNumberParserCache").With("ttx", time.Since(ttx)).Info("Loaded PhoneNumberData items", "items", len(PhoneNumberData))
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
