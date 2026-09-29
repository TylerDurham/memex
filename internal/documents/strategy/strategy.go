package strategy

type SkipDirectories map[string]struct{}

type IndexStrategy interface {
	SkipDirectories() SkipDirectories
}
