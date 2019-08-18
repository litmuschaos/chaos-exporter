package bdd_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChaos(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = Describe("BDD on chaos-exporter", func() {
	Context("Chaos Engine Liveliness test", func() {

		It("should be a chaosEngine", func() {
			app := "kubectl"
			arg1 := "get"
			arg2 := "chaosengine"
			arg3 := "-n"
			arg4 := "litmus"
			cmd := exec.Command(app, arg1, arg2, arg3, arg4)
			stdout, err := cmd.Output()

			if err != nil {
				println(err.Error())
				return
			}

			println(string(stdout))

			Expect(string(stdout)).To(MatchRegexp("engine-nginx"))

		})
	})
})
