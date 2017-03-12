package types

type AuthDriverType string

// AuthActionType allow or deny user
type AuthActionType string

func (a AuthActionType) Bool() bool {
	switch a {
	case AuthActionTypeAllow:
		return true
	case AuthActionTypeDeny:
		return false
	}
	return false
}

type AuthAllowDenyType bool
