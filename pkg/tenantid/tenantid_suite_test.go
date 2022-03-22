package tenantid_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTenantid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenantid Suite")
}
