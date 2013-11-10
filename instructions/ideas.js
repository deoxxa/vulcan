// Hello world!
{code: 200, message: "Hello, world!"}

// Proxy to upstreams
{upstreams: ["localhost:5000", "localhost:5001"]}

// Get upstreams from etcd
{upstreams: upstreams("upstreams/*")}

// Get upstreams from etcd matched by path
{upstreams: upstreams_match(request.path)}

// Basic rate limiting
{
    rates: {"$request.ip": "100 requests/second"}, 
    upstreams: ["localhost:5000", "localhost:5001"]
}

// Rate limit by client_ip
{
    rates: {"$request.client_ip": "100 requests/second"},
    upstreams: ["localhost:5000", "localhost:5001"]
}

// Rate limit by username
{
    rates: {"$request.username": "100 requests/second"},
    upstreams: ["localhost:5000", "localhost:5001"]
}

// Authenticate request against auth server and rate limit by account id
auth = get(
    "localhost:5000/auth", 
    params={
        username: request.username, 
        password: request.password
    })

if (auth['code'] != 200) {
    result = auth
} else {
    result = {
        rates: {
            result['account-id']: "100 requests/second"
        }
    }
}

// Authenticate request against auth server, using cache
auth = cache_get("auth", request.username + request.password)
if (auth == null) {
    auth = get(
        "localhost:5000/auth", 
        params={
            username: request.username, 
            password: request.password
        })
}



if (auth['code'] != 200) {
    result = auth
} else {
    result = {
        rates: {
            result['account-id']: "100 requests/second"
        }
    }
}
