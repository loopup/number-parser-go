# number-parser-go

[![Build Status](https://dev.azure.com/loopup/Cloud-Telephony/_apis/build/status%2Floopup.number-parser-go?branchName=main)](https://dev.azure.com/loopup/Cloud-Telephony/_build/latest?definitionId=300&branchName=main)
![](https://img.shields.io/github/v/tag/loopup/number-parser-go)
![](https://img.shields.io/azure-devops/coverage/loopup/number-parser-go/300)
![](https://img.shields.io/azure-devops/tests/loopup/number-parser-go/300)


```sh
go get -u github.com/loopup/number-parser-go@latest
```

## Usage

```go
    // You must pass a number free of decorations to the FindNumberForE164()
    r1 := FindNumberDataForE164(SanitizeNumber("+1 (415) 777-1234"))
    // Equivalent to the..
    r2 := FindNumberDataForE164("14157771234")
```

## References
- The [country code information](https://en.wikipedia.org/wiki/List_of_telephone_country_codes) is from Wikipedia.
- Detailed number data is partially sourced from carriers.