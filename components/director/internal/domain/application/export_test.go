package application

func NewConverter() *converter {
	return &converter{}
}

func (r *Resolver) SetConverter(converter ApplicationConverter) {
	r.converter = converter
}

