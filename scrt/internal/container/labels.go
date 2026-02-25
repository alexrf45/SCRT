package container

// Label keys used to identify and filter SCRT containers.
const (
	LabelAuthor  = "author"
	LabelProject = "scrt.project"
	LabelVersion = "scrt.version"

	// AuthorValue is the default author label value for filtering.
	AuthorValue = "fr3d"
)

// Labels builds the standard SCRT label set for a container.
func Labels(project, version string) map[string]string {
	return map[string]string{
		LabelAuthor:  AuthorValue,
		LabelProject: project,
		LabelVersion: version,
	}
}

// FilterLabels returns the label filter string for listing SCRT containers.
func FilterLabels() string {
	return LabelAuthor + "=" + AuthorValue
}
