package boxconfig

import "chatterbox/boxq/types"

type AuthConfiguration struct {
	Driver      types.AuthDriverType
	AuthNoMatch types.AuthActionType
	ACLNoMatch  types.AuthActionType
}
