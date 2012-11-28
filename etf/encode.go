package etf

type EncodeError struct {
	Msg string
}

func (err EncodeError) Error() string {
	return errPrefix + "encode error: " + err.Msg
}
