package tenantid

// Indicates that no tenant matches the provided identifier
type TenantNotFoundError struct {
	msg string
}

func (e *TenantNotFoundError) Error() string { return e.msg }
