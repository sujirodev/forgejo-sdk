// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMyUser(t *testing.T) {
	log.Println("== TestMyUser ==")

	var expectedAvatarURL, expectedHTMLURL string
	if os.Getenv("CI") != "" {
		expectedAvatarURL = "http://forgejo:3000/avatars/90e9f0102fc2832d69ae59a1214601c0"
		expectedHTMLURL = "http://forgejo:3000/test01"
	} else {
		expectedAvatarURL = "http://localhost:3000/avatars/90e9f0102fc2832d69ae59a1214601c0"
		expectedHTMLURL = "http://localhost:3000/test01"
	}

	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	assert.EqualValues(t, 1, user.ID)
	assert.EqualValues(t, "test01", user.UserName)
	assert.EqualValues(t, "test01@forgejo.org", user.Email)
	assert.EqualValues(t, "", user.FullName)
	assert.EqualValues(t, expectedAvatarURL, user.AvatarURL)
	assert.EqualValues(t, expectedHTMLURL, user.HTMLURL)
	assert.True(t, user.IsAdmin)
}

func TestUserApp(t *testing.T) {
	log.Println("== TestUserApp ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	result, _, err := c.ListAccessTokens(user.UserName, ListAccessTokensOptions{})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	// TODO: the gitea-admin name for the token is hardcoded in forgejo itself, until it's changed this will need to do
	assert.EqualValues(t, "gitea-admin", result[0].Name)

	t1, _, err := c.CreateAccessToken(user.UserName, CreateAccessTokenOption{Name: "TestCreateAccessToken", Scopes: []AccessTokenScope{AccessTokenScopeRepositoryRead}})
	require.NoError(t, err)
	assert.EqualValues(t, "TestCreateAccessToken", t1.Name)
	result, _, _ = c.ListAccessTokens(user.UserName, ListAccessTokensOptions{})
	assert.Len(t, result, 2)

	_, err = c.DeleteAccessToken(user.UserName, t1.ID)
	require.NoError(t, err)
	result, _, _ = c.ListAccessTokens(user.UserName, ListAccessTokensOptions{})
	assert.Len(t, result, 1)
}

func TestUserSearch(t *testing.T) {
	log.Println("== TestUserSearch ==")
	c := newTestClient()

	createTestUser(t, "tu1", c)
	createTestUser(t, "eatIt_2", c)
	createTestUser(t, "thirdIs3", c)
	createTestUser(t, "advancedUser", c)
	createTestUser(t, "1n2n3n", c)
	createTestUser(t, "otherIt", c)

	ul, _, err := c.SearchUsers(SearchUsersOption{KeyWord: "other"})
	require.NoError(t, err)
	assert.Len(t, ul, 1)

	ul, _, err = c.SearchUsers(SearchUsersOption{KeyWord: "notInTESTcase"})
	require.NoError(t, err)
	assert.Empty(t, ul)

	ul, _, err = c.SearchUsers(SearchUsersOption{KeyWord: "It"})
	require.NoError(t, err)
	assert.Len(t, ul, 2)
}

func TestUserFollow(t *testing.T) {
	log.Println("== TestUserFollow ==")
	c := newTestClient()
	me, _, _ := c.GetMyUserInfo()

	uA := "uFollow_A"
	uB := "uFollow_B"
	uC := "uFollow_C"
	createTestUser(t, uA, c)
	createTestUser(t, uB, c)
	createTestUser(t, uC, c)

	// A follow ME
	// B follow C & ME
	// C follow A & B & ME
	c.sudo = uA
	_, err := c.Follow(me.UserName)
	require.NoError(t, err)
	c.sudo = uB
	_, err = c.Follow(me.UserName)
	require.NoError(t, err)
	_, err = c.Follow(uC)
	require.NoError(t, err)
	c.sudo = uC
	_, err = c.Follow(me.UserName)
	require.NoError(t, err)
	_, err = c.Follow(uA)
	require.NoError(t, err)
	_, err = c.Follow(uB)
	require.NoError(t, err)

	// C unfollow me
	_, err = c.Unfollow(me.UserName)
	require.NoError(t, err)

	// ListMyFollowers of me
	c.sudo = ""
	f, _, err := c.ListMyFollowers(ListFollowersOptions{})
	require.NoError(t, err)
	assert.Len(t, f, 2)

	// ListFollowers of A
	f, _, err = c.ListFollowers(uA, ListFollowersOptions{})
	require.NoError(t, err)
	assert.Len(t, f, 1)

	// ListMyFollowing of me
	f, _, err = c.ListMyFollowing(ListFollowingOptions{})
	require.NoError(t, err)
	assert.Empty(t, f)

	// ListFollowing of A
	f, _, err = c.ListFollowing(uA, ListFollowingOptions{})
	require.NoError(t, err)
	assert.Len(t, f, 1)
	assert.EqualValues(t, me.ID, f[0].ID)

	isFollow, _ := c.IsFollowing(uA)
	assert.False(t, isFollow)
	isFollow, _ = c.IsUserFollowing(uB, uC)
	assert.True(t, isFollow)
}

func TestUserEmail(t *testing.T) {
	log.Println("== TestUserEmail ==")
	c := newTestClient()
	userN := "TestUserEmail"
	createTestUser(t, userN, c)
	c.sudo = userN

	// ListEmails
	el, _, err := c.ListEmails(ListEmailsOptions{})
	require.NoError(t, err)
	assert.Len(t, el, 1)
	assert.EqualValues(t, "TestUserEmail@forgejo.org", el[0].Email)
	assert.True(t, el[0].Primary)

	// AddEmail
	mails := []string{"wow@mail.send", "speed@mail.me"}
	el, _, err = c.AddEmail(CreateEmailOption{Emails: mails})
	require.NoError(t, err)
	assert.Len(t, el, 3)
	_, _, err = c.AddEmail(CreateEmailOption{Emails: []string{mails[1]}})
	require.Error(t, err)
	el, _, err = c.ListEmails(ListEmailsOptions{})
	require.NoError(t, err)
	assert.Len(t, el, 3)

	// DeleteEmail
	_, err = c.DeleteEmail(DeleteEmailOption{Emails: []string{mails[1]}})
	require.NoError(t, err)
	_, err = c.DeleteEmail(DeleteEmailOption{Emails: []string{"imaginary@e.de"}})
	require.Error(t, err)

	el, _, err = c.ListEmails(ListEmailsOptions{})
	require.NoError(t, err)
	assert.Len(t, el, 2)
	_, err = c.DeleteEmail(DeleteEmailOption{Emails: []string{mails[0]}})
	require.NoError(t, err)
}

func TestGetUserByID(t *testing.T) {
	log.Println("== TestGetUserByID ==")
	c := newTestClient()

	user1 := createTestUser(t, "user1", c)
	user2 := createTestUser(t, "user2", c)

	r1, _, err := c.GetUserByID(user1.ID)
	require.NoError(t, err)
	assert.NotNil(t, r1)
	assert.Equal(t, user1.UserName, r1.UserName)

	r2, _, err := c.GetUserByID(user2.ID)
	require.NoError(t, err)
	assert.NotNil(t, r2)
	assert.Equal(t, user2.UserName, r2.UserName)

	r3, _, err := c.GetUserByID(42)
	require.Error(t, err)
	assert.Nil(t, r3)

	r4, _, err := c.GetUserByID(-1)
	require.Error(t, err)
	assert.Nil(t, r4)
}

func createTestUser(t *testing.T, username string, client *Client) *User {
	user, _, _ := client.GetUserInfo(username)
	if user.ID != 0 {
		return user
	}
	user, _, err := client.AdminCreateUser(CreateUserOption{Username: username, Password: username + "!1234", Email: username + "@forgejo.org", MustChangePassword: OptionalBool(false), SendNotify: false})
	require.NoError(t, err)
	return user
}

// userToStringSlice return string slice based on UserName of users
func userToStringSlice(users []*User) []string {
	result := make([]string, 0, len(users))
	for i := range users {
		result = append(result, users[i].UserName)
	}
	return result
}
