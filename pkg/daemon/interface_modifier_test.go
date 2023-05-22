package daemon_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/daemon"
)

var _ = Context("interface modifier", func() {
	It("can string", func() {
		for i := 0; i < len(daemon.InterfaceModifiersStrings); i++ {
			mod := daemon.InterfaceModifier(1 << i)

			Expect(mod.String()).To(Equal(daemon.InterfaceModifiersStrings[i]))
		}
	})

	It("can strings", func() {
		mod := daemon.InterfaceModifier(math.MaxInt)
		for i := 0; i < len(daemon.InterfaceModifiersStrings); i++ {
			Expect(mod.Strings()).To(Equal(daemon.InterfaceModifiersStrings))
		}
	})

	It("can check if set", func() {
		mod := daemon.InterfaceModifiedPeers

		Expect(mod.Is(daemon.InterfaceModifiedName)).To(BeFalse())
		Expect(mod.Is(daemon.InterfaceModifiedPeers)).To(BeTrue())
	})
})
