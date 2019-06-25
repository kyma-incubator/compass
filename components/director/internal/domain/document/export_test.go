package document

func (r *Resolver) SetConverter(converter DocumentConverter) {
	r.converter = converter
}
