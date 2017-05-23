// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	api "code.gitea.io/sdk/gitea"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/routers/api/v1/convert"
)

// ListEmails list all the emails of mine
func ListEmails(ctx *context.APIContext) {
	// swagger:route GET /emails userGetEmails
	//
	// List current user emails.
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       200: EmailList
	//       500: error

	emails, err := models.GetEmailAddresses(ctx.User.ID)
	if err != nil {
		ctx.Error(500, "GetEmailAddresses", err)
		return
	}
	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert.ToEmail(emails[i])
	}
	ctx.JSON(200, &apiEmails)
}

// AddEmail add email for me
func AddEmail(ctx *context.APIContext, form api.CreateEmailOption) {
	// swagger:route POST /emails userAddEmails
	//
	// Add emails to current user.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       201: EmailList
	//       422: validationError
	//       500: error

	if len(form.Emails) == 0 {
		ctx.Status(422)
		return
	}

	emails := make([]*models.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &models.EmailAddress{
			UID:         ctx.User.ID,
			Email:       form.Emails[i],
			IsActivated: !setting.Service.RegisterEmailConfirm,
		}
	}

	if err := models.AddEmailAddresses(emails); err != nil {
		if models.IsErrEmailAlreadyUsed(err) {
			ctx.Error(422, "", "Email address has been used: "+err.(models.ErrEmailAlreadyUsed).Email)
		} else {
			ctx.Error(500, "AddEmailAddresses", err)
		}
		return
	}

	apiEmails := make([]*api.Email, len(emails))
	for i := range emails {
		apiEmails[i] = convert.ToEmail(emails[i])
	}
	ctx.JSON(201, &apiEmails)
}

// DeleteEmail delete email
func DeleteEmail(ctx *context.APIContext, form api.CreateEmailOption) {
	// swagger:route DELETE	 /emails userDeleteEmails
	//
	// Remove email address to current user.
	//
	//     Consumes:
	//     - application/json
	//
	//     Produces:
	//     - application/json
	//
	//     Responses:
	//       204: empty
	//       500: error

	if len(form.Emails) == 0 {
		ctx.Status(204)
		return
	}

	emails := make([]*models.EmailAddress, len(form.Emails))
	for i := range form.Emails {
		emails[i] = &models.EmailAddress{
			Email: form.Emails[i],
			UID:   ctx.User.ID,
		}
	}

	if err := models.DeleteEmailAddresses(emails); err != nil {
		ctx.Error(500, "DeleteEmailAddresses", err)
		return
	}
	ctx.Status(204)
}
