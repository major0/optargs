module github.com/major0/optargs/goarg

go 1.23.4

require github.com/major0/optargs v0.0.0

// Local development - replace with parent module
replace github.com/major0/optargs => ../

// Test mode - use upstream alexflint/go-arg for compatibility testing
// This configuration allows us to test against the real upstream implementation
