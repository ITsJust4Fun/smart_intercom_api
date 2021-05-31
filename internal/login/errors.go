package login

type WrongPasswordError struct{}

func (m *WrongPasswordError) Error() string {
	return "wrong password"
}
