package flows

// ID names a transcript routing context (base flow).
type ID string

const (
	Root       ID = "root"
	Select     ID = "select"
	SelectFile ID = "select_file"
)
