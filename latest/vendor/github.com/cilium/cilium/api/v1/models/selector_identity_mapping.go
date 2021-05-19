// Code generated by go-swagger; DO NOT EDIT.

// Copyright 2017-2021 Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// SelectorIdentityMapping mapping of selector to identities which match it
//
// swagger:model SelectorIdentityMapping
type SelectorIdentityMapping struct {

	// identities mapping to this selector
	Identities []int64 `json:"identities"`

	// string form of selector
	Selector string `json:"selector,omitempty"`

	// number of users of this selector in the cache
	Users int64 `json:"users,omitempty"`
}

// Validate validates this selector identity mapping
func (m *SelectorIdentityMapping) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *SelectorIdentityMapping) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SelectorIdentityMapping) UnmarshalBinary(b []byte) error {
	var res SelectorIdentityMapping
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
