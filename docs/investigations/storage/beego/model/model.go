package model

//TODO

type Application struct {
	ID          string
	Tenant      string
	Name        string
	Description string
	Labels      string // JSON
	Apis        APIPage
	Documents   DocumentPage
}

type API struct {
	ID        string
	TargetURL string
}

type Document struct {
	ID     string
	Title  string
	Format string
	Data   string
}

type ApplicationPage struct {
	Data     []Application
	PageInfo PageInfo
}

type APIPage struct {
	Data     []API
	PageInfo PageInfo
}

type DocumentPage struct {
	Data     []Document
	PageInfo PageInfo
}

type PageRequest struct {
	AfterCursor string
	PageSize    int
}

type PageInfo struct {
	TotalCount             int
	HasNextPage            bool
	StartCursor, EndCursor string
}

type Filer struct {
	Labels map[string]string
}
