package goapi

type TagEntry struct {
	Name        string
	Description string
}

type Tags []TagEntry

func (t *Tags) Set(name string) int {
	for index, el := range *t {
		if el.Name == name {
			return index
		}
	}

	*t = append(*t, TagEntry{Name: name})

	return len(*t) - 1
}

func (t *Tags) SetDescription(name string, description string) {
	index := t.Set(name)

	(*t)[index].Description = description
}
