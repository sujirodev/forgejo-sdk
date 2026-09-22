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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifications(t *testing.T) {
	log.Println("== TestNotifications ==")

	c := newTestClient()

	// Dedicated accounts. This test used to make the shared admin account
	// (test01) the notified user, and its last block asserted that account
	// held exactly 2 notifications. That 2 never came from anything this
	// test did: run alone against a fresh instance the account holds 0, and
	// the assertion only ever passed because of what the rest of the suite
	// happened to leave in test01's inbox. Every account here is created by
	// this test, so every count below is one this test caused.
	//
	// createTestUser needs the admin account, so all three are created
	// before the first SetSudo.
	user1 := createTestUser(t, uniqueName(t, "notifyowner"), c)
	user2 := createTestUser(t, uniqueName(t, "notifywatch"), c)
	reader := createTestUser(t, uniqueName(t, "notifyread"), c)
	t.Cleanup(func() { c.SetSudo("") })

	// create 2 repos
	c.SetSudo(user1.UserName)
	repoA, err := createTestRepo(t, uniqueName(t, "notifA"), c)
	require.NoError(t, err)

	c.sudo = user2.UserName
	repoB, err := createTestRepo(t, uniqueName(t, "notifB"), c)
	require.NoError(t, err)
	_, err = c.WatchRepo(user1.UserName, repoA.Name)
	c.sudo = user1.UserName
	require.NoError(t, err)

	c.sudo = user2.UserName
	notifications, _, err := c.ReadNotifications(MarkNotificationOptions{})
	require.NoError(t, err)
	assert.Empty(t, notifications)
	count, _, err := c.CheckNotifications()
	assert.Equal(t, int64(0), count)
	require.NoError(t, err)
	c.sudo = user1.UserName
	_, _, err = c.CreateIssue(repoA.Owner.UserName, repoA.Name, CreateIssueOption{Title: "A Issue", Closed: false})
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(repoB.Owner.UserName, repoB.Name, CreateIssueOption{Title: "B Issue", Closed: false})
	require.NoError(t, err)

	// CheckNotifications of user2. Forgejo delivers notifications
	// asynchronously; poll instead of sleeping a fixed amount, which stops
	// being enough under load.
	c.sudo = user2.UserName
	eventually(t, func() bool {
		count, _, err = c.CheckNotifications()
		return err == nil && count == 2
	})
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
	c.sudo = user1.UserName
	_, _, err = c.EditIssue(repoB.Owner.UserName, repoB.Name, issue.Index, EditIssueOption{State: &iState})
	require.NoError(t, err)

	c.sudo = user2.UserName
	eventually(t, func() bool {
		count, _, err = c.CheckNotifications()
		return err == nil && count == 1
	})
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

	// The remaining routes -- ReadNotifications with no filter, the Status
	// filters, and ReadNotification's pinned/unread forms -- need an account
	// holding exactly two notifications. Build one instead of borrowing
	// whatever the shared admin account happened to accumulate: `reader`
	// watches repoA, its inbox is drained, and then user1 opens exactly two
	// issues there.
	c.sudo = reader.UserName
	_, err = c.WatchRepo(user1.UserName, repoA.Name)
	require.NoError(t, err)
	_, _, err = c.ReadNotifications(MarkNotificationOptions{})
	require.NoError(t, err)

	c.sudo = user1.UserName
	_, _, err = c.CreateIssue(repoA.Owner.UserName, repoA.Name, CreateIssueOption{Title: "C Issue", Closed: false})
	require.NoError(t, err)
	_, _, err = c.CreateIssue(repoA.Owner.UserName, repoA.Name, CreateIssueOption{Title: "D Issue", Closed: false})
	require.NoError(t, err)

	c.sudo = reader.UserName
	eventually(t, func() bool {
		n, _, cErr := c.CheckNotifications()
		return cErr == nil && n == 2
	})
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
