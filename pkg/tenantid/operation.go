package tenantid

import "context"

type operationKeyType int

const (
	operationKey operationKeyType = iota
)

//nolint:golint,deadcode,unused
func setOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationKey, operation)
}

func operation(ctx context.Context) string {
	value := ctx.Value(operationKey)

	if value == nil {
		return "unknown"
	}

	return value.(string)
}
