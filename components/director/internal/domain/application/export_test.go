package application

func (r *Resolver) SetConverter(converter ApplicationConverter) {
	r.appConverter = converter
}

