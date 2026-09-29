package strategy

type SkipDirectories map[string]struct{}
type Extensions map[string]struct{}

type IndexStrategy interface {
	SkipDirectories() SkipDirectories
	Extensions() Extensions
}
