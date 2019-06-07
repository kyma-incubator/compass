package pagination

type Page struct {
	StartCursor string
	EndCursor string
	HasNextPage bool
}
