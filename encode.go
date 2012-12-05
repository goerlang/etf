package etf

type EncodeError struct {
	Msg string
}

func (err EncodeError) Error() string {
	return "etf: encode error: " + err.Msg
}
