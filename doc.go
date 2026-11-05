// Package checkly is a Go client for the Checkly public API.
//
// NewClient builds a Client, configured by options:
//
//	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
//	if err != nil {
//		log.Fatal(err)
//	}
//	check, err := client.GetAPICheck(ctx, "a1b2c3d4-0000-4000-8000-000000000001")
//
// WithAuthFromEnv reads the API key and account ID from CHECKLY_API_KEY and
// CHECKLY_ACCOUNT_ID; WithAuth takes them as arguments. A client without auth
// can only call the public endpoints, GetRuntime and GetStaticIPs.
//
// Each check and monitor type has its own type and methods, such as APICheck
// with CreateAPICheck, GetAPICheck, UpdateAPICheck and DeleteAPICheck.
//
// A method returns an *APIError when the API answers with a status the method
// does not accept. It matches the sentinel errors, such as ErrNotFound and
// ErrRateLimited, with errors.Is. A request answered with HTTP 429 is retried
// a few times, waiting as the API asks for up to a minute and while the
// context's deadline allows; UploadCodeBundle, which streams its body, is not
// retried.
package checkly
