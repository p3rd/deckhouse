/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"testing"

	"github.com/Masterminds/semver/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/fake"
	"github.com/iancoleman/strcase"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/deckhouse/go_lib/dependency"
	"github.com/deckhouse/deckhouse/go_lib/dependency/cr"
	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Modules :: deckhouse :: hooks :: check deckhouse release ::", func() {
	f := HookExecutionConfigInit(`{
"deckhouse":{
  "releaseChannel": "Stable",
  "internal":{
	"currentReleaseImageName":"dev-registry.deckhouse.io/sys/deckhouse-oss/dev:test"}
  }
}`, `{}`)
	f.RegisterCRD("deckhouse.io", "v1alpha1", "DeckhouseRelease", false)

	dependency.TestDC.CRClient = cr.NewClientMock(GinkgoT())
	Context("No new deckhouse image", func() {
		BeforeEach(func() {
			dependency.TestDC.CRClient.ImageMock.Return(&fake.FakeImage{LayersStub: func() ([]v1.Layer, error) {
				return []v1.Layer{&fakeLayer{}, &fakeLayer{Body: `{"version": "v1.25.3"}`}}, nil
			}}, nil)
			f.KubeStateSet("")
			f.BindingContexts.Set(f.GenerateScheduleContext("* * * * *"))
			f.RunHook()
		})
		It("Release should be created", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-3").Exists()).To(BeTrue())
			Expect(f.KubernetesGlobalResource("DeckhouseRelease", "v1-25-3").Field("spec.version").String()).To(BeEquivalentTo("v1.25.3"))
		})
	})
})

type fakeLayer struct {
	v1.Layer
	Body string
}

func (fl fakeLayer) Uncompressed() (io.ReadCloser, error) {
	result := bytes.NewBuffer(nil)

	if fl.Body == "" {
		return ioutil.NopCloser(result), nil
	}

	// returns tar file with content
	// {"version": "v1.25.3"}
	body := json.RawMessage(fl.Body)
	hdr := &tar.Header{
		Name: "version.json",
		Mode: 0600,
		Size: int64(len(body)),
	}
	wr := tar.NewWriter(result)
	_ = wr.WriteHeader(hdr)
	_, _ = wr.Write(body)
	_ = wr.Close()

	return ioutil.NopCloser(result), nil
}

func (fl fakeLayer) Size() (int64, error) {
	return int64(len(fl.Body)), nil
}

func TestSort(t *testing.T) {
	s1 := deckhouseReleaseUpdate{
		Version: semver.MustParse("v1.24.0"),
	}
	s2 := deckhouseReleaseUpdate{
		Version: semver.MustParse("v1.24.1"),
	}
	s3 := deckhouseReleaseUpdate{
		Version: semver.MustParse("v1.24.2"),
	}
	s4 := deckhouseReleaseUpdate{
		Version: semver.MustParse("v1.24.3"),
	}
	s5 := deckhouseReleaseUpdate{
		Version: semver.MustParse("v1.24.4"),
	}

	releases := []deckhouseReleaseUpdate{s3, s4, s1, s5, s2}
	sort.Sort(sort.Reverse(byVersion(releases)))

	for i, rl := range releases {
		if rl.Version.String() != "1.24."+strconv.FormatInt(int64(4-i), 10) {
			t.Fail()
		}
	}

}

func TestKebabCase(t *testing.T) {
	cases := map[string]string{
		"Alpha":       "alpha",
		"Beta":        "beta",
		"EarlyAccess": "early-access",
		"Stable":      "stable",
		"RockSolid":   "rock-solid",
	}

	for original, kebabed := range cases {
		result := strcase.ToKebab(original)

		assert.Equal(t, result, kebabed)
	}
}
