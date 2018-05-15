// Code generated by go-swagger; DO NOT EDIT.

package store

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware/dispatch/pkg/function-manager/gen/models"
)

// NewUpdateFunctionParams creates a new UpdateFunctionParams object
// no default values defined in spec.
func NewUpdateFunctionParams() UpdateFunctionParams {

	return UpdateFunctionParams{}
}

// UpdateFunctionParams contains all the bound params for the update function operation
// typically these are obtained from a http.Request
//
// swagger:parameters updateFunction
type UpdateFunctionParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*function object
	  Required: true
	  In: body
	*/
	Body *models.Function
	/*Name of function to work on
	  Required: true
	  Pattern: ^[\w\d\-]+$
	  In: path
	*/
	FunctionName string
	/*Filter based on tags
	  In: query
	  Collection Format: multi
	*/
	Tags []string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewUpdateFunctionParams() beforehand.
func (o *UpdateFunctionParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.Function
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("body", "body"))
			} else {
				res = append(res, errors.NewParseError("body", "body", "", err))
			}
		} else {

			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Body = &body
			}
		}
	} else {
		res = append(res, errors.Required("body", "body"))
	}
	rFunctionName, rhkFunctionName, _ := route.Params.GetOK("functionName")
	if err := o.bindFunctionName(rFunctionName, rhkFunctionName, route.Formats); err != nil {
		res = append(res, err)
	}

	qTags, qhkTags, _ := qs.GetOK("tags")
	if err := o.bindTags(qTags, qhkTags, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *UpdateFunctionParams) bindFunctionName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.FunctionName = raw

	if err := o.validateFunctionName(formats); err != nil {
		return err
	}

	return nil
}

func (o *UpdateFunctionParams) validateFunctionName(formats strfmt.Registry) error {

	if err := validate.Pattern("functionName", "path", o.FunctionName, `^[\w\d\-]+$`); err != nil {
		return err
	}

	return nil
}

func (o *UpdateFunctionParams) bindTags(rawData []string, hasKey bool, formats strfmt.Registry) error {

	// CollectionFormat: multi
	tagsIC := rawData

	if len(tagsIC) == 0 {
		return nil
	}

	var tagsIR []string
	for _, tagsIV := range tagsIC {
		tagsI := tagsIV

		tagsIR = append(tagsIR, tagsI)
	}

	o.Tags = tagsIR

	return nil
}
