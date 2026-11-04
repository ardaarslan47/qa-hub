package qahub

type Description struct {
	text string
}

func ParseDescription(raw string) Description {
	return Description{text: trimInvisible(raw)}
}

func (d Description) String() string {
	return d.text
}
