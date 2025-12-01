package models

// Shared constants for parameter types and TeamCity-specific raw values.
// Keep these in a central place to avoid string drift across packages.
const (
	// Terraform-visible parameter types
	ParamTypeText     = "text"
	ParamTypePassword = "password"

	// TeamCity REST API raw type for secure parameters
	SecureParamRawType = "password display='normal'"
)
