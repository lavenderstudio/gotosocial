// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"code.superseriousbusiness.org/gopkg/httputil"
	"code.superseriousbusiness.org/gopkg/httputil/binding"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/go-playground/form/v4"
)

// ParseFocus parses a media attachment focus parameters from incoming API string.
func ParseFocus(focus string) (focusx, focusy float32, errWithCode gtserror.WithCode) {
	if focus == "" {
		return
	}
	spl := strings.Split(focus, ",")
	if len(spl) != 2 {
		const text = "missing comma separator"
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	xStr := spl[0]
	yStr := spl[1]
	fx, err := strconv.ParseFloat(xStr, 32)
	if err != nil || fx > 1 || fx < -1 {
		text := fmt.Sprintf("invalid x focus: %s", xStr)
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	fy, err := strconv.ParseFloat(yStr, 32)
	if err != nil || fy > 1 || fy < -1 {
		text := fmt.Sprintf("invalid y focus: %s", xStr)
		errWithCode = gtserror.NewErrorBadRequest(
			errors.New(text),
			text,
		)
		return
	}
	focusx = float32(fx)
	focusy = float32(fy)
	return
}

// ParseDuration parses the given raw interface belonging
// the given fieldName as an integer duration.
func ParseDuration(rawI any, fieldName string) (*int, error) {
	var (
		asInteger int
		err       error
	)

	switch raw := rawI.(type) {
	case float64:
		// Submitted as JSON number
		// (casts to float64 by default).
		asInteger = int(raw)

	case string:
		// Submitted as JSON string or form field.
		asInteger, err = strconv.Atoi(raw)
		if err != nil {
			err = fmt.Errorf(
				"could not parse %s value %s as integer: %w",
				fieldName, raw, err,
			)
		}

	default:
		// Submitted as god-knows-what.
		err = fmt.Errorf(
			"could not parse %s type %T as integer",
			fieldName, rawI,
		)
	}

	if err != nil {
		return nil, err
	}

	return &asInteger, nil
}

// ParseNullableDuration is like ParseDuration, but
// for JSON values that may have been sent as `null`.
//
// IsSpecified should be checked and "true" on the
// given nullable before calling this function.
func ParseNullableDuration(
	nullable apimodel.Nullable[any],
	fieldName string,
) (*int, error) {
	if nullable.IsNull() {
		// Was specified as `null`,
		// return pointer to zero value.
		return util.Ptr(0), nil
	}

	rawI, err := nullable.Get()
	if err != nil {
		return nil, err
	}

	return ParseDuration(rawI, fieldName)
}

func parseFieldsAttributesFromJSON(jsonFieldsAttributes *map[string]apimodel.UpdateField) (*[]apimodel.UpdateField, error) {
	if jsonFieldsAttributes == nil {
		// Nothing set, nothing to do.
		return nil, nil
	}

	fieldsAttributes := make([]apimodel.UpdateField, 0, len(*jsonFieldsAttributes))
	for keyStr, updateField := range *jsonFieldsAttributes {
		key, err := strconv.Atoi(keyStr)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse fieldAttributes key %s to int: %w", keyStr, err)
		}

		fieldsAttributes = append(fieldsAttributes, apimodel.UpdateField{
			Key:   key,
			Name:  updateField.Name,
			Value: updateField.Value,
		})
	}

	// Sort slice by the key each field was submitted with.
	slices.SortFunc(fieldsAttributes, func(a, b apimodel.UpdateField) int {
		const k = +1
		switch {
		case a.Key > b.Key:
			return +k
		case a.Key < b.Key:
			return -k
		default:
			return 0
		}
	})

	return &fieldsAttributes, nil
}

// fieldsAttributesFormBinding satisfies
// httputil's binding.Binding interface.
//
// Should only be used specifically
// for multipart/form-data MIME type.
type fieldsAttributesFormBinding struct{}

func (fieldsAttributesFormBinding) Name() string {
	return "FieldsAttributes"
}

func (fieldsAttributesFormBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	// Change default namespace prefix
	// and suffix to allow correct parsing
	// of the field attributes.
	decoder := form.NewDecoder()
	decoder.SetNamespacePrefix("[")
	decoder.SetNamespaceSuffix("]")

	return decoder.Decode(obj, req.Form)
}

func ParseWithFieldsAttributes(
	c *httputil.Context,
	form apimodel.WithFieldsAttributes,
) (apimodel.WithFieldsAttributes, error) {
	// nolint
	maxMemory := int64(config.GetHTTPServerMaxMultipartMemory())

	switch ct := c.ContentType(); ct {
	case binding.MIMEJSON:
		// Bind with default json binding first.
		if err := binding.BindJSON(c, form, maxMemory); err != nil {
			return nil, err
		}

		// Now use custom form binding for
		// field attributes in the json data.
		fa, err := parseFieldsAttributesFromJSON(form.GetJSONFieldsAttributes())
		if err != nil {
			return nil, fmt.Errorf("custom json binding failed: %w", err)
		}
		form.SetFieldsAttributes(fa)

	case binding.MIMEPOSTForm:
		// Bind with default form binding first.
		if err := binding.BindForm(c, form, maxMemory); err != nil {
			return nil, err
		}

		// Now use custom form binding for
		// field attributes in the form data.
		if err := (fieldsAttributesFormBinding{}).Bind(c.R, form); err != nil {
			return nil, fmt.Errorf("custom form binding failed: %w", err)
		}

	case binding.MIMEMultipartPOSTForm:
		// Bind with default form binding first.
		if err := binding.BindFormMultipart(c, form, maxMemory); err != nil {
			return nil, err
		}

		// Now use custom form binding for
		// field attributes in the form data.
		if err := (fieldsAttributesFormBinding{}).Bind(c.R, form); err != nil {
			return nil, fmt.Errorf("custom form binding failed: %w", err)
		}

	default:
		err := fmt.Errorf(
			"content-type %s not supported for this endpoint; supported content-types are %s, %s, %s",
			ct, binding.MIMEJSON, binding.MIMEPOSTForm, binding.MIMEMultipartPOSTForm,
		)
		return nil, err
	}

	return form, nil
}
