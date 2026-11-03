package domain

type UserError string

func (e UserError) Error() string {
	return string(e)
}
