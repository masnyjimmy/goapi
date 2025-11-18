package goapi

type ServerEntry struct {
	Url         string
	Description string
}

type Servers []ServerEntry

func (s *Servers) Set(url string) int {
	for index, el := range *s {
		if el.Url == url {
			return index
		}
	}

	*s = append(*s, ServerEntry{Url: url})

	return len(*s) - 1
}

func (s *Servers) SetDescription(url string, description string) {
	index := s.Set(url)

	(*s)[index].Description = description
}
