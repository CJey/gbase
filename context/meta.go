package context

type sessionKey struct{}

// SetSession use to set real session name
func SetSession(ctx Context, session string) {
	ctx.Set(sessionKey{}, session)
}

// GetSession use to get session name if it has real session, otherwise return context name instead
func GetSession(ctx Context) string {
	if session, ok := ctx.Get(sessionKey{}); ok {
		return session.(string)
	}
	return ctx.Name()
}

// GetRealSession use to get real session name
func GetRealSession(ctx Context) string {
	return ctx.GetString(sessionKey{})
}
