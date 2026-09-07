package mapper

var stripScripts = Mapper{
	Name:     "strip-scripts",
	Accepts:  []Kind{KindHTML},
	Produces: KindSame,
	Transform: func(payload Payload) (Payload, error) {
		payload.Content = htmlScriptBlock.ReplaceAll(payload.Content, nil)
		return payload, nil
	},
}
