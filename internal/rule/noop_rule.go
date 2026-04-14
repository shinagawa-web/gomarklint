package rule

// CheckNoop is a no-op rule that immediately returns nil.
// Used only for benchmark investigation to test whether adding a 13th
// call site in collectErrors causes a regression independent of any
// rule logic.
func CheckNoop(_ string, _ []string, _ int) []LintError {
	return nil
}
