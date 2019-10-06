package self_test

import (
	"crypto/tls"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/lucas-clemente/quic-go/integrationtests/tools/testlog"
	"github.com/lucas-clemente/quic-go/internal/testdata"
)

const alpn = "quic-go integration tests"

func getTLSConfig() *tls.Config {
	conf := testdata.GetTLSConfig()
	conf.NextProtos = []string{alpn}
	return conf
}

func getTLSClientConfig() *tls.Config {
	return &tls.Config{
		RootCAs:    testdata.GetRootCA(),
		NextProtos: []string{alpn},
	}
}

func scaleDuration(d time.Duration) time.Duration {
	scaleFactor := 1
	if f, err := strconv.Atoi(os.Getenv("TIMESCALE_FACTOR")); err == nil { // parsing "" errors, so this works fine if the env is not set
		scaleFactor = f
	}
	Expect(scaleFactor).ToNot(BeZero())
	return time.Duration(scaleFactor) * d
}

func TestSelf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Self integration tests")
}

var _ = BeforeSuite(func() {
	rand.Seed(GinkgoRandomSeed())
})
