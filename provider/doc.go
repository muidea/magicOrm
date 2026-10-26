// Package provider is the application entry point for Local and Remote models.
// Use NewLocalProviderWithOptions for Go structs and NewRemoteProviderWithOptions
// for dynamic declarations. Both return Provider and use the models contracts.
// Subpackage remote defines dynamic declarations and values; helper converts
// representations. Neither subpackage provides network transport.
package provider
