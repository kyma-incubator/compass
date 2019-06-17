package runtime

func NewConverter() *converter {
	return &converter{}
}

func (r *Resolver) SetConverter(converter RuntimeConverter) {
	r.converter = converter
}

