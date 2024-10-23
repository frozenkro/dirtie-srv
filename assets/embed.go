package assets

import "embed"

//go:embed html
var AssetDir embed.FS

var ChangePasswordPageKey string = "html/changePasswordPage.html"
var ResetPwEmailKey string = "html/resetPwEmail.html"
