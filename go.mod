module github.com/loopup/number-parser-go

go 1.23.3

retract (
    v0.3.3 // Obsolete; buggy
    v0.3.2 // Obsolete
    v0.3.1 // Changed the file name
    v0.3.0 // Incorrect module name
)

require github.com/gocarina/gocsv v0.0.0-20240520201108-78e41c74b4b1
