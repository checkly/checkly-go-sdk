// Package checkly is a Go client for the Checkly public API.
//
// [NewClient] builds a [Client], configured by options:
//
//	client, err := checkly.NewClient(checkly.WithAuthFromEnv())
//	if err != nil {
//		log.Fatal(err)
//	}
//	check, err := client.GetAPICheck(ctx, "a1b2c3d4-0000-4000-8000-000000000001")
//
// [WithAuthFromEnv] reads the API key and account ID from CHECKLY_API_KEY and
// CHECKLY_ACCOUNT_ID; [WithAuth] takes them as arguments. A client without auth
// can only call the public endpoints, GetRuntime and GetStaticIPs.
//
// Each check and monitor type has its own type and methods, such as [APICheck]
// with CreateAPICheck, GetAPICheck, UpdateAPICheck and DeleteAPICheck.
// For an ID whose check type is not known, [Client.GetCheck] returns a [Check]
// and [Client.DeleteCheck] deletes it.
//
// # Errors, cancellation and retries
//
// A method returns an [*APIError] when the API answers with a status the
// method does not accept. It matches the sentinel errors, such as
// [ErrNotFound] and [ErrRateLimited], with errors.Is; getting a resource by a
// well-formed but unknown ID, for example, fails with an error matching
// ErrNotFound.
//
// Requests are sent with the context the method is given: when the context
// is cancelled or its deadline passes, the request and any wait before a
// retry are abandoned. A request answered with HTTP 429 is retried a few
// times, waiting as the API asks for up to a minute and while the context's
// deadline allows; UploadCodeBundle, which streams its body, is not retried.
// A GET, PUT or DELETE request that gets no response, for example because
// its connection was reset, is retried twice after short waits, also while
// the deadline allows, and so is PeekCodeBundle. A timeout, a DNS failure, a
// rejected certificate or a failure to connect is not retried, since a retry
// would only repeat it. Other POST requests are not retried, because a create may have
// happened even though its response was lost. The error of a request that
// was retried says how often: a re-sent DELETE, for example, gets a 404 if
// the first attempt succeeded.
//
// # Updates
//
// An Update method sends every settable field of the value it is given, so
// zero values are applied, except fields whose comment says they are left
// out of the request when nil, empty, zero or false. A left-out field keeps
// its current value, with these exceptions: fields an Update method names as
// reset to the API default; AlertChannelSubscriptions, described below; and
// fields nested in AlertSettings, in a group's APICheckDefaults, or in an
// alert channel's or status page component's configuration, which the API
// replaces together with the object they are in, as it does for the request
// of DNS monitors. To change some fields, start from the value a Get returns.
//
// Creates and updates of checks, monitors and check groups ask the API not
// to subscribe alert channels automatically, so a resource written with an
// empty AlertChannelSubscriptions has no alert channel subscriptions.
package checkly
