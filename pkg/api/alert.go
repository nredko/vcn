/*
 * Copyright (c) 2018-2020 vChain, Inc. All Rights Reserved.
 * This software is released under GPL3.
 * The full license information can be found under:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 */

package api

import (
	"fmt"

	"github.com/vchain-us/vcn/pkg/meta"
)

type alert struct {
	ArtifactHash     string   `json:"artifactHash,omitempty"`
	ArtifactMetaHash string   `json:"artifactMetaHash,omitempty"`
	Enabled          bool     `json:"enabled"`
	Metadata         Metadata `json:"metadata,omitempty"`
	UUID             string   `json:"uuid,omitempty"`
}

type alertNotification struct {
	AlertUUID string   `json:"alertUUID"`
	Metadata  Metadata `json:"metadata,omitempty"`
}

// CreateAlert creates a platform alert and returns its UUID.
func (u *User) CreateAlert(a Artifact, v BlockchainVerification, m Metadata) (uuid string, err error) {

	restError := new(Error)
	alertResponse := &alert{}
	r, err := newSling(u.token()).
		Post(meta.APIEndpoint("alert")).
		BodyJSON(alert{
			ArtifactHash:     a.Hash,
			ArtifactMetaHash: v.MetaHash(),
			Enabled:          true,
			Metadata:         m,
		}).Receive(alertResponse, restError)

	if err != nil {
		return
	}
	if r.StatusCode == 200 {
		uuid = alertResponse.UUID
	} else {
		err = fmt.Errorf("request failed: %s (%d)", restError.Message,
			restError.Status)
	}
	return
}

// ModifyAlert modifies the settings of an already existing alert.
func (u *User) ModifyAlert(uuid string, enabled bool, m *Metadata) error {

	restError := new(Error)
	alertResponse := &alert{}
	alertRequest := alert{
		Enabled: enabled,
	}
	if m != nil {
		alertRequest.Metadata = *m
	}
	r, err := newSling(u.token()).
		Patch(meta.APIEndpoint("alert")+"?uuid="+uuid).
		BodyJSON(alertRequest).Receive(alertResponse, restError)

	if err != nil {
		return err
	}

	switch r.StatusCode {
	case 200:
		return nil
	case 403:
		return fmt.Errorf("illegal alert access: %s", restError.Message)
	case 404:
		return fmt.Errorf(`no such alert found matching "%s"`, uuid)
	default:
		return fmt.Errorf("request failed: %s (%d)", restError.Message,
			restError.Status)
	}
}

func (u *User) alertMessage(uuid string, what string) (err error) {
	restError := new(Error)
	r, err := newSling(u.token()).
		Post(meta.APIEndpoint("alert/"+what)).
		BodyJSON(alertNotification{
			AlertUUID: uuid,
		}).Receive(nil, restError)

	if err != nil {
		return
	}

	switch r.StatusCode {
	case 200:
		return nil
	case 403:
		return fmt.Errorf("illegal alert access: %s", restError.Message)
	case 404:
		return fmt.Errorf(`no such alert found matching "%s"`, uuid)
	case 412:
		return fmt.Errorf(`notification already triggered for "%s"`, uuid)
	default:
		return fmt.Errorf("request failed: %s (%d)", restError.Message,
			restError.Status)
	}
}

// PingAlert sends a ping for the given alert _uuid_.
// Once the first ping goes through, the platform starts a server-side watcher and will trigger a notification
// after some amount of time if no further pings for _uuid_ are received.
func (u *User) PingAlert(uuid string) error {
	return u.alertMessage(uuid, "ping")
}

// TriggerAlert triggers a notification immediately for the given alert _uuid_.
func (u *User) TriggerAlert(uuid string) error {
	return u.alertMessage(uuid, "notify")
}