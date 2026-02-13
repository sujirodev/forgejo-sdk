// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifications(t *testing.T) {
	log.Println("== TestNotifications ==")

	// init user2
	c := newTestClient()

	user1, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	user2 := createTestUser(t, "notify2", c)

	// create 2 repos
	repoA, err := createTestRepo(t, "TestNotifications_A", c)
	require.NoError(t, err)

	c.sudo = user2.UserName
	repoB, err := createTestRepo(t, "TestNotifications_B", c)
	require.NoError(t, err)
	_, err = c.WatchRepo(user1.UserName, repoA.Name)
	c.sudo = ""
	require.NoError(t, err)

	c.sudo = user2.UserName
	notifications, _, err := c.ReadNotifications(MarkNotificationOptions{})
	require.NoError(t, err)
	assert.Empty(t, notifications)
	count, _, err := c.CheckNotifications()
	assert.Equal(t, int64(0), count)
	require.NoError(t, err)
	c.sudo = ""
	_, _, err = c.CreateIssue(repoA.Owner.UserName, repoA.Name, CreateIssueOption{Title: "A Issue", Closed: false})
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(repoB.Owner.UserName, repoB.Name, CreateIssueOption{Title: "B Issue", Closed: false})
	require.NoError(t, err)
	time.Sleep(time.Second * 5)

	// CheckNotifications of user2
	c.sudo = user2.UserName
	count, _, err = c.CheckNotifications()
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// ListNotifications
	nList, _, err := c.ListNotifications(ListNotificationOptions{})
	require.NoError(t, err)
	assert.Len(t, nList, 2)
	for _, n := range nList {
		assert.True(t, n.Unread)
		assert.Equal(t, NotifySubjectType("Issue"), n.Subject.Type)
		assert.Equal(t, NotifySubjectOpen, nList[0].Subject.State)
		assert.Equal(t, NotifySubjectOpen, nList[1].Subject.State)
		switch n.Subject.Title {
		case "A Issue":
			assert.Equal(t, repoA.Name, n.Repository.Name)
		case "B Issue":
			assert.Equal(t, repoB.Name, n.Repository.Name)
		default:
			require.Error(t, fmt.Errorf("ListNotifications returned a Issue witch should not"))
		}
	}

	// ListRepoNotifications
	nList, _, err = c.ListRepoNotifications(repoA.Owner.UserName, repoA.Name, ListNotificationOptions{})
	require.NoError(t, err)
	assert.Len(t, nList, 1)
	assert.Equal(t, "A Issue", nList[0].Subject.Title)
	// ReadRepoNotifications
	notifications, _, err = c.ReadRepoNotifications(repoA.Owner.UserName, repoA.Name, MarkNotificationOptions{})
	require.NoError(t, err)
	assert.Len(t, notifications, 1)

	// GetThread
	n, _, err := c.GetNotification(nList[0].ID)
	require.NoError(t, err)
	assert.False(t, n.Unread)
	assert.Equal(t, "A Issue", n.Subject.Title)

	// ReadNotifications
	notifications, _, err = c.ReadNotifications(MarkNotificationOptions{})
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
	nList, _, err = c.ListNotifications(ListNotificationOptions{})
	require.NoError(t, err)
	assert.Empty(t, nList)

	// ReadThread
	iState := StateClosed
	c.sudo = ""
	_, _, err = c.EditIssue(repoB.Owner.UserName, repoB.Name, issue.Index, EditIssueOption{State: &iState})
	require.NoError(t, err)
	time.Sleep(time.Second * 5)

	c.sudo = user2.UserName
	nList, _, err = c.ListNotifications(ListNotificationOptions{})
	require.NoError(t, err)
	count, _, err = c.CheckNotifications()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	if assert.Len(t, nList, 1) {
		assert.Equal(t, NotifySubjectClosed, nList[0].Subject.State)
		notification, _, err := c.ReadNotification(nList[0].ID)
		require.NoError(t, err)
		assert.Equal(t, notification.ID, nList[0].ID)
	}

	c.sudo = ""
	notifications, _, err = c.ReadNotifications(MarkNotificationOptions{})
	require.NoError(t, err)
	assert.Len(t, notifications, 2)
	nList, _, err = c.ListNotifications(ListNotificationOptions{Status: []NotifyStatus{NotifyStatusRead}})
	require.NoError(t, err)
	if assert.Len(t, nList, 2) {
		notification, _, err := c.ReadNotification(nList[0].ID, NotifyStatusPinned)
		assert.Equal(t, notification.ID, nList[0].ID)
		require.NoError(t, err)

		notification, _, err = c.ReadNotification(nList[1].ID, NotifyStatusUnread)
		assert.Equal(t, notification.ID, nList[1].ID)
		require.NoError(t, err)
	}
	nList, _, err = c.ListNotifications(ListNotificationOptions{Status: []NotifyStatus{NotifyStatusPinned, NotifyStatusUnread}})
	require.NoError(t, err)
	if assert.Len(t, nList, 2) {
		assert.Equal(t, NotifySubjectOpen, nList[0].Subject.State)
		assert.Equal(t, NotifySubjectOpen, nList[1].Subject.State)
	}
}
