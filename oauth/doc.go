// Package oauth provides utilities to wrap OAuth authentication flow.
//
// A Manager wraps all the logic and configuration of the OAuth flow:
//   - Paths (default redirect, login handler, callback handler)
//   - OAuth config
//   - Authenticator and Authorizer to translate access tokens to User
//
// When using the Middleware, the User is flow is:
//   - Check session state
//     - If session still valid, allow the request to proceed
//   - Redirect user to the login page to kickstart auth flow, recording origin URL
//   - Redirect user to upstream OAuth service, which may prompt user to accept
//   - Upstream redirects to callback handler with an access token
//   - The Authenticator determines which User corresponds to the access token
//   - The Authorizer determines if the User is allowed to be logged in.
//   - The user is redirected to their origin URL.
package oauth
