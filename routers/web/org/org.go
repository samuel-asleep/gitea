// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2018 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package org

import (
"errors"
"net/http"

"code.gitea.io/gitea/models/db"
"code.gitea.io/gitea/models/organization"
user_model "code.gitea.io/gitea/models/user"
"code.gitea.io/gitea/modules/log"
"code.gitea.io/gitea/modules/setting"
"code.gitea.io/gitea/modules/templates"
"code.gitea.io/gitea/modules/web"
"code.gitea.io/gitea/services/context"
"code.gitea.io/gitea/services/forms"
)

const (
// tplCreateOrg template path for create organization
tplCreateOrg templates.TplName = "org/create"
)

// Create render the page for create organization
func Create(ctx *context.Context) {
ctx.Data["Title"] = ctx.Tr("new_org")
if !ctx.Doer.CanCreateOrganization() {
ctx.ServerError("Not allowed", errors.New(ctx.Locale.TrString("org.form.create_org_not_allowed")))
return
}

ctx.Data["visibility"] = setting.Service.DefaultOrgVisibilityMode
ctx.Data["repo_admin_change_team_access"] = true

ctx.HTML(http.StatusOK, tplCreateOrg)
}

// CreateSubOrg renders the page to create a subgroup under an existing organization.
func CreateSubOrg(ctx *context.Context) {
ctx.Data["Title"] = ctx.Tr("org.create_sub_org")
if !ctx.Doer.CanCreateOrganization() {
ctx.ServerError("Not allowed", errors.New(ctx.Locale.TrString("org.form.create_org_not_allowed")))
return
}

parentOrg := ctx.Org.Organization
isOwner, err := organization.IsOrganizationOwner(ctx, parentOrg.ID, ctx.Doer.ID)
if err != nil {
ctx.ServerError("IsOrganizationOwner", err)
return
}
if !isOwner {
ctx.HTTPError(http.StatusForbidden)
return
}

ctx.Data["ParentOrg"] = parentOrg
ctx.Data["ParentOrgID"] = parentOrg.ID
ctx.Data["visibility"] = parentOrg.Visibility
ctx.Data["repo_admin_change_team_access"] = true

ctx.HTML(http.StatusOK, tplCreateOrg)
}

// CreatePost response for create organization
func CreatePost(ctx *context.Context) {
form := *web.GetForm(ctx).(*forms.CreateOrgForm)
ctx.Data["Title"] = ctx.Tr("new_org")

if !ctx.Doer.CanCreateOrganization() {
ctx.ServerError("Not allowed", errors.New(ctx.Locale.TrString("org.form.create_org_not_allowed")))
return
}

if ctx.HasError() {
ctx.HTML(http.StatusOK, tplCreateOrg)
return
}

org := &organization.Organization{
Name:                      form.OrgName,
IsActive:                  true,
Type:                      user_model.UserTypeOrganization,
Visibility:                form.Visibility,
RepoAdminChangeTeamAccess: form.RepoAdminChangeTeamAccess,
ParentID:                  form.ParentOrgID,
}

// When creating a subgroup, re-display the parent org info on error.
if form.ParentOrgID != 0 {
parentOrg, err := organization.GetOrgByID(ctx, form.ParentOrgID)
if err == nil {
ctx.Data["ParentOrg"] = parentOrg
ctx.Data["ParentOrgID"] = parentOrg.ID
}
}

if err := organization.CreateOrganization(ctx, org, ctx.Doer); err != nil {
ctx.Data["Err_OrgName"] = true
switch {
case user_model.IsErrUserAlreadyExist(err):
ctx.RenderWithErrDeprecated(ctx.Tr("form.org_name_been_taken"), tplCreateOrg, &form)
case db.IsErrNameReserved(err):
ctx.RenderWithErrDeprecated(ctx.Tr("org.form.name_reserved", err.(db.ErrNameReserved).Name), tplCreateOrg, &form)
case db.IsErrNamePatternNotAllowed(err):
ctx.RenderWithErrDeprecated(ctx.Tr("org.form.name_pattern_not_allowed", err.(db.ErrNamePatternNotAllowed).Pattern), tplCreateOrg, &form)
case organization.IsErrUserNotAllowedCreateOrg(err):
ctx.RenderWithErrDeprecated(ctx.Tr("org.form.create_org_not_allowed"), tplCreateOrg, &form)
default:
ctx.ServerError("CreateOrganization", err)
}
return
}
log.Trace("Organization created: %s", org.Name)

ctx.Redirect(org.AsUser().DashboardLink())
}
