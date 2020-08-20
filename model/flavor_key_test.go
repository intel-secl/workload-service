/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	commLog "intel/isecl/lib/common/v3/log"
	"intel/isecl/lib/flavor/v3"
	"testing"
)

type FlavorKeyInfo struct {
	Flavor    flavor.Image `json:"flavor"`
	Signature string       `json:"signature"`
	Key       []byte       `json:"key"`
}

var log = commLog.GetDefaultLogger()

func TestFlavorKeyDeserialize(t *testing.T) {
	log.Trace("model/flavor_key_test:TestFlavorKeyDeserialize() Entering")
	defer log.Trace("model/flavor_key_test:TestFlavorKeyDeserialize() Leaving")
	raw := `{
		"flavor": {
		  "meta": {
			"id": "c10cc1f5-7c32-4d24-9a1e-17cb5f6b125d",
			"description": {
			  "flavor_part": "IMAGE",
			  "label": "label"
			}
		  },
		  "encryption": {
			"encryption_required": true,
			"key_url": "https://kbs.server.com:443/v1/keys/ecee021e-9669-4e53-9224-8880fb4e4080/transfer",
			"digest": "C6CR+CsFIqfxEgkoe4hWVQyBwJ71amh1TwVQBo4TWa0="
		  }
		},
		"signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8k",
		"key": "6WNBU9HfOF0hVynmGaTOG3wSc4R0hlxWe0hOXW3WAwNwjpiyo5/OExrfYkIhHwRlBWPhkpKfAy6gJ9AOzPr3VJFCuU+chtBNtPcgoXIxa02vJSiLGmYOXzaVDUZD1Ht7ND2rrVQB8cqjYC/0FXfpUu65/KFhbvTE1liXrmfmfFbrgQGoDqIA8rjuMtZTTLDuIZiOQZR9jdhjysDzEkZUDglqIbN+fIY/8c1Bsge7E26YEpBWqPNTwmuNmFwHRr4f6HSGmbUwVUietwGjRkbp/CSMIiWyLI99hvuQ/D5Tc1jEN4wvGPTvzc9MpecJ2at0BHRTx4MpEQ+u3XWW3zy1xA=="
	  }`
	var fk FlavorKeyInfo
	json.Unmarshal([]byte(raw), &fk)
	assert.Equal(t, "c10cc1f5-7c32-4d24-9a1e-17cb5f6b125d", fk.Flavor.Meta.ID)
	assert.Equal(t, "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8k", fk.Signature)
	assert.NotEqual(t, []uint8([]byte(nil)), fk.Key)
}
