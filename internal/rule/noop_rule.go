package rule

// CheckNoop* are no-op rules that immediately return nil.
// Used only for benchmark investigation to test compiler effects
// of varying numbers of if-blocks in collectErrors.

func CheckNoop1(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop2(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop3(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop4(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop5(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop6(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop7(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop8(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop9(_ string, _ []string, _ int) []LintError  { return nil }
func CheckNoop10(_ string, _ []string, _ int) []LintError { return nil }
func CheckNoop11(_ string, _ []string, _ int) []LintError { return nil }
func CheckNoop12(_ string, _ []string, _ int) []LintError { return nil }
func CheckNoop13(_ string, _ []string, _ int) []LintError { return nil }
