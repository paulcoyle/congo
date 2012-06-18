# What Does it Do?

Congo provides a small but useful facility to add a chain of responsibility to
http handlers (called "handler actions").  For those familiar  with Ruby on
Rails this is similar to before filters.

During the execution of the chain, a context is passed to each action and is
passed automatically to any html templates being rendered as the template
data.  Along the way, one can augment the context with custom structures (so
long as they implement `congo.Context`) to provide additional data or
functionality to templates.

The handler action chain will halt whenever an action returns a "response".
There are a handful of defined responses that represent basic things:
template rendering, redirecting (302) and returning 404.  Currently there is
no facility to add custom responses but this is certainly a good candidate for
a feature addition.

# Example

## Defining Handler Action Chains

Say you want to have a set of handlers that each must load a secure cookie
from the request and, depending on the content (if the user ID is present or
whatever) either redirect the user to an authentication page or display some
content.

We'll start by defining an augmented context:

    type SecureCookieContext struct {
      congo.Context
      Cookie xyz.SecureCookie // just some fictitious package
    }

So that subsequent actions and, possibly, templates can access the secure
cookies, we are augmenting the base `congo.Context` with `SecureCookieContext`.
In order to apply this, we need to then define the action that will do this:

    func addSecureCookieContext(c congo.Context) (congo.Context, interface{}) {
      newContext := &SecureCookieContext{c, globalCookieStruct}
      return newContext, nil // return nil for a response to allow the chain to continue
    }

Since we want this to be applied for all handlers, we can start by defining a
"root" handler whose first action is to augment the context with our
`SecureCookieContext`:

    rootHandler := congo.NewHandler().Actions(addSecureCookieContext)

Now we can add another context and action pair for augmenting the
`SecureCookieContext` with some user information and checking that the user is
logged in (simply checking if their ID exists in the cookie).

    type AuthContext struct {
      *SecureCookieContext
    }

    func (c *AuthContext) LoggedIn() bool {
      loggedIn := false
      // logic for checking cookie here
      return loggedIn
    }

    func addAuthContext(c congo.Context) (congo.Context, interface{}) {
      cookieContext, ok := c.(*SecureCookieContext)
      if !ok {
        panic("Expected inbound context to be *SecureCookieContext")
      }

      newContext := &AuthContext{cookieContext}
      return newContext, nil
    }

    func requireLoggedIn(c congo.Context) (congo.Context, interface{}) {
      authContext, ok := c.(*AuthContext)
      if !ok {
        panic("Expected inbound context to be *AuthContext")
      }

      if authContext.LoggedIn() {
        return authContext, nil // the user is logged in, allow the chain to continue
      } else {
        return authContext, &congo.RedirectResponse{"/login"}
      }
    }

Let's say we want to have the `AuthContext` available to all actions by
default.  We can just change our root handler as so:

    rootHandler := congo.NewHandler().Actions(addSecureCookieContext, addAuthContext)

Now we can branch off the root handler to create some handlers for specific
pages.  Let's define one for `/account` and one for `/login` where `/account`
can only be accessed by a logged in user.

    loginHandler := rootHandler.Copy().Actions(loginHandler)
    accntHandler := rootHandler.Copy().Actions(requireLoggedIn, accountHandler)

Here we've made copies of the root handler which retain its previously
defined `addSecureCookieContext` and `addAuthContext` actions and then
appended other actions on to the end of the chains.

We can now define our terminal actions for `/login` and `/account`:

    func loginHandler(c congo.Context) (congo.Context, interface{}) {
      return c, &congo.RenderResponse("login", "layout")
    }

    func accntHandler(c congo.Context) (congo.Context, interface{}) {
      return c, &congo.RenderResponse("account", "layout")
    }

Great, now the handler action chains are all set up.  Now we need to define
how templates are rendered when we return a `congo.RenderResponse`.


## Handling Templates

Currently, congo can be given a default template store which is applied to all
handlers or a template store can be defined on specific handlers.  Generally,
it's easiest to have a single store that is applied to all handlers for the
sake of simplicity.

Template stores are simply pointers to `template.Template` from the core
`html/template` package.  Templates are referred to by the names defined within
these stores when returning a `congo.RenderResponse`.

The layout concept is also borrowed from other web frameworks.  Templates are
rendered then provided to the layout which uses the output of the template
somewhere in it's content.  From `examples/basic.go`:

    store, _ := template.New("layout").Parse("LAYOUT START (SECURE: {{.IsSecure}})\n{{.Content}}\nLAYOUT END")
    store.New("home").Parse("----HOME----")
    store.New("lion").Parse("----LION----\n----{{.Sound}}----")

This defines three templates: "layout", "home" and "lion".  As you can see,
the "layout" template refers to `{{.Content}}` which is where the template
results are made available to layouts.

So for this example we can do something similar (of course, you'll likely want
to use that newfangled HTML technology):

    store, _ := template.New("layout").Parse("HEADER\n{{.Content}}\nFOOTER")
    store.New("login").Parse("You must login to access your account")
    store.New("account").Parse("Welcome to your account!")


## Putting it Together

Now we just need to wire up our handlers to the template store and the http
server.  First, we'll define our template store as the default for all
handlers.  Note that this should appear *before* any handler creation.

    congo.SetDefaultTemplateStore(store)
    // handler creation from earlier in the example comes after this point

Finally, we can adapt our handlers to the `net/http` library's
`http.HandlerFunc` and start the server like so:

    http.HandleFunc("/login", congo.MuxHandler(loginHandler))
    http.HandleFunc("/account", congo.MuxHandler(accountHandler))
    http.ListenAndServe(":8080", nil)
