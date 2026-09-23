// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ActivityPubObject is a decoded ActivityStreams 2.0 JSON-LD document, such
// as an Actor (Application, Person or Repository), an Activity, or an
// OrderedCollection (an outbox). Forgejo's swagger only documents a fixed
// "@context" field for these payloads -- the actual shape (id, type,
// preferredUsername, inbox, outbox, publicKey, orderedItems, ...) is
// open-ended ActivityStreams JSON-LD defined by the federation modules
// (forgejo.org/modules/forgefed), not by the REST API's own schema -- so the
// SDK exposes the decoded document as a generic map rather than guessing at
// undocumented fields.
type ActivityPubObject map[string]interface{}

// GetActivityPubActor returns the instance's Actor (Application) document.
//
// This is the one federation route Forgejo serves unsigned on every version,
// so that a peer can bootstrap. Every other route in this file speaks to the
// caller as a remote federated server and needs an HTTP Signature: see
// UseActivityPubSignature.
func (c *Client) GetActivityPubActor() (*ActivityPubObject, *Response, error) {
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", "/activitypub/actor", nil, nil, obj)
	return obj, resp, err
}

// SendActivityPubActorInbox delivers an Activity to the instance actor's inbox.
func (c *Client) SendActivityPubActorInbox(activity ActivityPubObject) (*Response, error) {
	body, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", "/activitypub/actor/inbox", jsonHeader, bytes.NewReader(body))
	return resp, err
}

// GetActivityPubActorOutbox returns the instance actor's outbox, an
// ActivityStreams OrderedCollection. The instance actor's outbox is always
// empty.
func (c *Client) GetActivityPubActorOutbox() (*ActivityPubObject, *Response, error) {
	// The route was added in Forgejo 14.0.0: routers/api/v1/api.go registers
	// /activitypub/actor/outbox from v14.0.1 on, and not in v12.0.1 or
	// v13.0.1. A live 11.0.16 instance answers 404.
	if err := c.checkServerVersionGreaterThanOrEqual(version14_0_0); err != nil {
		return nil, nil, err
	}
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", "/activitypub/actor/outbox", nil, nil, obj)
	return obj, resp, err
}

// GetActivityPubRepository returns the Repository actor for the repo with the given ID.
func (c *Client) GetActivityPubRepository(repoID int64) (*ActivityPubObject, *Response, error) {
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/repository-id/%d", repoID), nil, nil, obj)
	return obj, resp, err
}

// SendActivityPubRepositoryInbox delivers an Activity to a repository's inbox.
func (c *Client) SendActivityPubRepositoryInbox(repoID int64, activity ActivityPubObject) (*Response, error) {
	body, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/activitypub/repository-id/%d/inbox", repoID), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// GetActivityPubRepositoryOutbox returns a repository's outbox, an
// ActivityStreams OrderedCollection.
func (c *Client) GetActivityPubRepositoryOutbox(repoID int64) (*ActivityPubObject, *Response, error) {
	// Added in Forgejo 14.0.0, alongside /activitypub/actor/outbox.
	if err := c.checkServerVersionGreaterThanOrEqual(version14_0_0); err != nil {
		return nil, nil, err
	}
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/repository-id/%d/outbox", repoID), nil, nil, obj)
	return obj, resp, err
}

// GetActivityPubPerson returns the Person actor for the user with the given ID.
func (c *Client) GetActivityPubPerson(userID int64) (*ActivityPubObject, *Response, error) {
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/user-id/%d", userID), nil, nil, obj)
	return obj, resp, err
}

// GetActivityPubPersonActivityNote returns the Note object of a specific
// activity recorded for the user.
func (c *Client) GetActivityPubPersonActivityNote(userID, activityID int64) (*ActivityPubObject, *Response, error) {
	// The route was added in Forgejo 13.0.0: routers/api/v1/api.go registers
	// /activitypub/user-id/{id}/activities/{id} from v13.0.1 on, and not in
	// v12.0.1. A live 11.0.16 instance answers 404.
	if err := c.checkServerVersionGreaterThanOrEqual(version13_0_0); err != nil {
		return nil, nil, err
	}
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/user-id/%d/activities/%d", userID, activityID), nil, nil, obj)
	return obj, resp, err
}

// GetActivityPubPersonActivity returns a specific activity recorded for the user.
func (c *Client) GetActivityPubPersonActivity(userID, activityID int64) (*ActivityPubObject, *Response, error) {
	// Added in Forgejo 13.0.0, alongside the Note route above.
	if err := c.checkServerVersionGreaterThanOrEqual(version13_0_0); err != nil {
		return nil, nil, err
	}
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/user-id/%d/activities/%d/activity", userID, activityID), nil, nil, obj)
	return obj, resp, err
}

// SendActivityPubPersonInbox delivers an Activity to a user's inbox.
func (c *Client) SendActivityPubPersonInbox(userID int64, activity ActivityPubObject) (*Response, error) {
	body, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/activitypub/user-id/%d/inbox", userID), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// GetActivityPubPersonOutbox lists the user's recorded activity (their outbox).
func (c *Client) GetActivityPubPersonOutbox(userID int64) (*ActivityPubObject, *Response, error) {
	// Added in Forgejo 13.0.0, alongside the activities routes above.
	if err := c.checkServerVersionGreaterThanOrEqual(version13_0_0); err != nil {
		return nil, nil, err
	}
	obj := new(ActivityPubObject)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/activitypub/user-id/%d/outbox", userID), nil, nil, obj)
	return obj, resp, err
}
