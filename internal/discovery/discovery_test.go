package discovery

func ids(bs []Browser) []string {
	out := []string{}
	for _, b := range bs {
		out = append(out, b.ID+"/"+string(b.Kind))
	}
	return out
}
