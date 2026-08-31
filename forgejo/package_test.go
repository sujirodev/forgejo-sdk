// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2023 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create an org with a single package for testing purposes
func createTestPackage(_ *testing.T, c *Client) error {
	_, _ = c.DeletePackage("PackageOrg", "generic", "MyPackage", "v1")
	_, _ = c.DeleteOrg("PackageOrg")
	_, _, _ = c.CreateOrg(CreateOrgOption{Name: "PackageOrg"})

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	reader := bytes.NewReader([]byte("Hello world!"))

	url := fmt.Sprintf("%s/api/packages/PackageOrg/generic/MyPackage/v1/file1.txt", os.Getenv("FORGEJO_SDK_TEST_URL"))
	req, err := http.NewRequest(http.MethodPut, url, reader)
	if err != nil {
		log.Println(err)
		return err
	}

	req.SetBasicAuth(os.Getenv("FORGEJO_SDK_TEST_USERNAME"), os.Getenv("FORGEJO_SDK_TEST_PASSWORD"))
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func TestListPackages(t *testing.T) {
	log.Println("== TestListPackages ==")
	c := newTestClient()
	err := createTestPackage(t, c)
	require.NoError(t, err)

	t.Run("Test without type", func(t *testing.T) {
		packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
			ListOptions: ListOptions{
				Page:     1,
				PageSize: 1000,
			},
		})
		require.NoError(t, err)
		assert.Len(t, packagesList, 1)
	})
	t.Run("Test with found", func(t *testing.T) {
		packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
			ListOptions: ListOptions{
				Page:     1,
				PageSize: 1000,
			},
			Type: PackageTypeGeneric,
		})
		require.NoError(t, err)
		assert.Len(t, packagesList, 1)
	})
	t.Run("Test with not found", func(t *testing.T) {
		packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
			ListOptions: ListOptions{
				Page:     1,
				PageSize: 1000,
			},
			Type: PackageTypeContainer,
		})
		require.NoError(t, err)
		assert.Len(t, packagesList, 0)
	})
	t.Run("Test with name query found", func(t *testing.T) {
		packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
			ListOptions: ListOptions{
				Page:     1,
				PageSize: 1000,
			},
			Q:    "My",
			Type: PackageTypeAll,
		})
		require.NoError(t, err)
		assert.Len(t, packagesList, 1)
	})
	t.Run("Test with name query not found", func(t *testing.T) {
		packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
			ListOptions: ListOptions{
				Page:     1,
				PageSize: 1000,
			},
			Q:    "NotMy",
			Type: PackageTypeAll,
		})
		require.NoError(t, err)
		assert.Len(t, packagesList, 0)
	})
}

func TestGetPackage(t *testing.T) {
	log.Println("== TestGetPackage ==")
	c := newTestClient()
	err := createTestPackage(t, c)
	require.NoError(t, err)

	pkg, _, err := c.GetPackage("PackageOrg", "generic", "MyPackage", "v1")
	require.NoError(t, err)
	assert.NotNil(t, pkg)
	assert.Equal(t, "MyPackage", pkg.Name)
	assert.Equal(t, "v1", pkg.Version)
	assert.NotEmpty(t, pkg.CreatedAt)
}

func TestDeletePackage(t *testing.T) {
	log.Println("== TestDeletePackage ==")
	c := newTestClient()
	err := createTestPackage(t, c)
	require.NoError(t, err)

	_, err = c.DeletePackage("PackageOrg", "generic", "MyPackage", "v1")
	require.NoError(t, err)

	// no packages should be listed following deletion
	packagesList, _, err := c.ListPackages("PackageOrg", ListPackagesOptions{
		ListOptions: ListOptions{
			Page:     1,
			PageSize: 1000,
		},
	})
	require.NoError(t, err)
	assert.Empty(t, packagesList)
}

func TestListPackageFiles(t *testing.T) {
	log.Println("== TestListPackageFiles ==")
	c := newTestClient()
	err := createTestPackage(t, c)
	require.NoError(t, err)

	packageFiles, _, err := c.ListPackageFiles("PackageOrg", "generic", "MyPackage", "v1")
	require.NoError(t, err)
	assert.Len(t, packageFiles, 1)
	assert.Equal(t, "file1.txt", packageFiles[0].Name)
}
