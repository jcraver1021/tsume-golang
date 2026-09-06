package mapper

var stripStyles = Mapper{
	Name:     "strip-styles",
	Accepts:  []Kind{KindHTML},
	Produces: KindSame,
	Transform: func(payload Payload) (Payload, error) {
		payload.Content = htmlStyleBlock.ReplaceAll(payload.Content, nil)
		return payload, nil
	},
}
