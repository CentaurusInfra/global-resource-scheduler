package app

var (
	// Version is ieg service current version
	Version   string
	commitID  string
	buildDate string
	goVersion string

	// Info produce environment info
	Info = ProdInfo{
		Version:   Version,
		CommitID:  commitID,
		BuildDate: buildDate,
		GoVersion: goVersion,
	}
)

// ProdInfo is the Produce Environment basic information
type ProdInfo struct {
	Version   string `json:"version" description:"-"`
	CommitID  string `json:"commit_id" description:"-"`
	BuildDate string `json:"build_date" description:"-"`
	GoVersion string `json:"go_version" description:"-"`
}
