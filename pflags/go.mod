module github.com/major0/optargs/pflags

go 1.23.4

require github.com/major0/optargs/goarg v0.0.0

// Local development - replace with local goarg module
replace github.com/major0/optargs/goarg => ../goarg

// Transitive dependency on optargs through goarg
// goarg/go.mod handles optargs dependency
