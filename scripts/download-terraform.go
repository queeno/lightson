package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

const terraformReleaseUrl = "https://releases.hashicorp.com/terraform/index.json"

type terraformRelease struct {
	Name     string `json:"name"`
	Versions map[string]struct {
		Name             string           `json:"name"`
		Version          string           `json:"version"`
		Shasums          string           `json:"shasums"`
		ShasumsSignature string           `json:"shasums_signature"`
		Builds           []terraformBuild `json:"builds"`
	} `json:"versions"`
}

type terraformBuild struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Os       string `json:"os"`
	Arch     string `json:"arch"`
	Filename string `json:"filename"`
	Url      string `json:"url"`
}

func returnTerraformUrlLinuxAmd64(builds []terraformBuild) (string, error) {
	for _, b := range builds {
		if b.Os == "linux" && b.Arch == "amd64" {
			return b.Url, nil
		}
	}
	return "", errors.Errorf("couldn't find a valid url for `linux:amd64`")
}

func downloadTerraformReleases(out io.Writer) (*terraformRelease, error) {
	res, err := http.Get(terraformReleaseUrl)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintf(out, "terraform releases downloaded\n")

	var tfR *terraformRelease
	err = json.Unmarshal(body, &tfR)
	if err != nil {
		return nil, err
	}

	return tfR, nil
}

func returnLatestTerraformVersion(tfR *terraformRelease) (string, error) {
	var versions version.Collection
	for k := range tfR.Versions {
		v, err := version.NewVersion(k)
		if err != nil {
			return "", err
		}
		versions = append(versions, v)
	}

	sort.Sort(sort.Reverse(versions))

	if len(versions) == 0 {
		return "", errors.Errorf("no terraform versions found")
	}

	return versions[0].String(), nil
}

func downloadTerraformReleaseFromUrl(out io.Writer, url string) (string, error) {
	_, _ = fmt.Fprintf(out, "download file from url: %s\n", url)

	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = res.Body }()

	zipf, err := ioutil.TempFile(os.TempDir(), "terraform-*.zip")
	if err != nil {
		return "", err
	}
	defer func() { _ = zipf.Close() }()

	_, err = io.Copy(zipf, res.Body)
	if err != nil {
		return "", err
	}
	return zipf.Name(), nil
}


func downloadTerraform(out io.Writer) error {
	tfR, err := downloadTerraformReleases(out)
	v, err := returnLatestTerraformVersion(tfR)

	_, _ = fmt.Fprintf(out, "the latest terraform release is: %s\n", v)

	builds := tfR.Versions[v].Builds

	if len(builds) == 0 {
		return errors.Errorf("no builds found for version %s", v)
	}

	url, err := returnTerraformUrlLinuxAmd64(builds)
	if err != nil {
		return err
	}

	filePath, err := downloadTerraformReleaseFromUrl(out, url)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", filePath)

	return nil
}

func main() {
	var out io.Writer

	if v := os.Getenv("DT_LOG"); v != "" {
		out = os.Stdout
	} else {
		out = ioutil.Discard
	}

	if err := downloadTerraform(out); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		os.Exit(1)
	}
}
