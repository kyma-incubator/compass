package repo

// OrderByDir is a type encapsulating the ORDER BY direction
type OrderByDir string

const (
	// AscOrderBy defines ascending order
	AscOrderBy OrderByDir = "ASC"
	// DescOrderBy defines descending order
	DescOrderBy OrderByDir = "DESC"
)

// OrderBy type that wraps the information about the ordering column and direction
type OrderBy struct {
	Field string
	Dir   OrderByDir
}

// NewAscOrderBy returns wrapping type for ascending order for a given column (field)
func NewAscOrderBy(field string) OrderBy {
	return OrderBy{
		Field: field,
		Dir:   AscOrderBy,
	}
}

// NewDescOrderBy returns wrapping type for descending orderd for a given column (field)
func NewDescOrderBy(field string) OrderBy {
	return OrderBy{
		Field: field,
		Dir:   DescOrderBy,
	}
}

// OrderByParams is a wrapping type for slice of OrderBy types
type OrderByParams []OrderBy

// NoOrderBy represents default ordering (no order specified)
var NoOrderBy = OrderByParams{}
