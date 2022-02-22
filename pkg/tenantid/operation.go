package tenantid

import "context"

type operationKeyType int

const (
	operationKey = iota
)

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
