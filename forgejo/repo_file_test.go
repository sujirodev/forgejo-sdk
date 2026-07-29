// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/base64"
	"log"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedLicenseTemplate = `MIT License

Copyright (c) [REDACTED]

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the
following conditions:

The quick brown fox jumps over the lazy dog and definitely did not read this license before agreeing
to its terms.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO
EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE
USE OR OTHER DEALINGS IN THE SOFTWARE.
`

func TestFileCreateUpdateGet(t *testing.T) {
	log.Println("== TestFileCRUD ==")
	c := newTestClient()

	repo, err := createTestRepo(t, "ChangeFiles", c)
	require.NoError(t, err)
	assert.NotNil(t, repo)

	raw, _, err := c.GetFile(repo.Owner.UserName, repo.Name, "main", "README.md")
	require.NoError(t, err)
	assert.Equal(t, "IyBDaGFuZ2VGaWxlcwoKQSB0ZXN0IFJlcG86IENoYW5nZUZpbGVz", base64.StdEncoding.EncodeToString(raw))

	testFileName := "A+#&ä"
	newFile, _, err := c.CreateFile(repo.Owner.UserName, repo.Name, testFileName, CreateFileOptions{
		FileOptions: FileOptions{
			Message: "create file " + testFileName,
		},
		Content: "ZmlsZUEK",
	})
	require.NoError(t, err)
	raw, _, _ = c.GetFile(repo.Owner.UserName, repo.Name, "main", testFileName)
	assert.Equal(t, "ZmlsZUEK", base64.StdEncoding.EncodeToString(raw))

	updatedFile, _, err := c.UpdateFile(repo.Owner.UserName, repo.Name, testFileName, UpdateFileOptions{
		FileOptions: FileOptions{
			Message: "add a new line",
		},
		SHA:     newFile.Content.SHA,
		Content: "ZmlsZUEKCmFuZCBhIG5ldyBsaW5lCg==",
	})
	require.NoError(t, err)
	assert.NotNil(t, updatedFile)

	file, _, err := c.GetContents(repo.Owner.UserName, repo.Name, "main", testFileName)
	require.NoError(t, err)
	assert.Equal(t, updatedFile.Content.SHA, file.SHA)
	assert.Equal(t, &updatedFile.Content.Content, &file.Content)

	_, err = c.DeleteFile(repo.Owner.UserName, repo.Name, testFileName, DeleteFileOptions{
		FileOptions: FileOptions{
			Message: "Delete File " + testFileName,
		},
		SHA: updatedFile.Content.SHA,
	})
	require.NoError(t, err)
	_, resp, err := c.GetFile(repo.Owner.UserName, repo.Name, "main", testFileName)
	require.Error(t, err)
	assert.Equal(t, "The target couldn't be found.", err.Error())
	assert.Equal(t, 404, resp.StatusCode)

	licence, _, err := c.GetContents(repo.Owner.UserName, repo.Name, "", "LICENSE")
	require.NoError(t, err)
	licenceRaw, _, err := c.GetFile(repo.Owner.UserName, repo.Name, "", "LICENSE")
	require.NoError(t, err)
	testContent := "diff --git a/LICENSE b/LICENSE\nindex 8086dc7..f402aed 100644\n--- a/LICENSE\n+++ b/LICENSE\n@@ -8,8 +8,8 @@ without limitation the rights to use, copy, modify, merge, publish, distribute,\n copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the\n following conditions:\n \n-The above copyright notice and this permission notice shall be included in all copies or substantial\n-portions of the Software.\n+The quick brown fox jumps over the lazy dog and definitely did not read this license before agreeing\n+to its terms.\n \n THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT\n LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO\n"
	updatedFile, _, err = c.DiffPatchFile(repo.Owner.UserName, repo.Name, DiffPatchFileOptions{
		FileOptions: FileOptions{
			Message:       "DiffPatch",
			BranchName:    "main",
			NewBranchName: "diffpatch-a+/&licence",
		},
		SHA:     licence.SHA,
		Content: testContent,
	})
	require.NoError(t, err)
	assert.NotNil(t, updatedFile)
	licenceRawNew, _, err := c.GetFile(repo.Owner.UserName, repo.Name, "diffpatch-a+/&licence", "LICENSE")
	require.NoError(t, err)
	assert.NotNil(t, licence)
	assert.False(t, bytes.Equal(licenceRaw, licenceRawNew))
	licenceTextNew := string(licenceRawNew)
	assert.Equal(t, expectedLicenseTemplate, regexp.MustCompile(`(?m)^Copyright \(c\) .*$`).ReplaceAllString(licenceTextNew, "Copyright (c) [REDACTED]"))

	licence, _, err = c.GetContents(repo.Owner.UserName, repo.Name, "", "LICENSE")
	require.NoError(t, err)
	licenceRaw, _, err = c.GetFile(repo.Owner.UserName, repo.Name, "", "LICENSE")
	require.NoError(t, err)
	testContent = "Tk9USElORyBJUyBIRVJFIEFOWU1PUkUKSUYgWU9VIExJS0UgVE8gRklORCBTT01FVEhJTkcKV0FJVCBGT1IgVEhFIEZVVFVSRQo="
	updatedFile, _, err = c.UpdateFile(repo.Owner.UserName, repo.Name, "LICENSE", UpdateFileOptions{
		FileOptions: FileOptions{
			Message:       "Overwrite",
			BranchName:    "main",
			NewBranchName: "overwrite-a+/&licence",
		},
		SHA:     licence.SHA,
		Content: testContent,
	})
	require.NoError(t, err)
	assert.NotNil(t, updatedFile)
	licenceRawNew, _, err = c.GetFile(repo.Owner.UserName, repo.Name, "overwrite-a+/&licence", "LICENSE")
	require.NoError(t, err)
	assert.NotNil(t, licence)
	assert.False(t, bytes.Equal(licenceRaw, licenceRawNew))
	assert.Equal(t, testContent, base64.StdEncoding.EncodeToString(licenceRawNew))

	// ListContents in root dir of default branch
	dir, resp, err := c.ListContents(repo.Owner.UserName, repo.Name, "", "")
	require.NoError(t, err)
	assert.Len(t, dir, 3)
	assert.NotNil(t, resp)

	// ListContents in not existing dir of default branch
	_, resp, err = c.ListContents(repo.Owner.UserName, repo.Name, "", "/hehe/")
	require.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	// ListContents in root dir of not existing branch
	_, resp, err = c.ListContents(repo.Owner.UserName, repo.Name, "no-ref-at-all", "")
	require.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// ListContents try to get file as dir
	dir, resp, err = c.ListContents(repo.Owner.UserName, repo.Name, "", "LICENSE")
	if assert.Error(t, err) {
		assert.Equal(t, "expect directory, got file", err.Error())
	}
	assert.Nil(t, dir)
	assert.Equal(t, 200, resp.StatusCode)

	// GetContents try to get dir as file
	file, resp, err = c.GetContents(repo.Owner.UserName, repo.Name, "", "")
	if assert.Error(t, err) {
		assert.Equal(t, "expect file, got directory", err.Error())
	}
	assert.Nil(t, file)
	assert.Equal(t, 200, resp.StatusCode)
}
