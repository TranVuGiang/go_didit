# go_didit

Go client library for the [Didit](https://didit.me) KYC verification API.

## Installation

```bash
go get github.com/TranVuGiang/go_didit
```

## Requirements

- Go 1.22+
- Didit API key

## Usage

### Create a client

```go
import godidit "github.com/TranVuGiang/go_didit"

client, err := godidit.NewClient("your-api-key")
if err != nil {
    log.Fatal(err)
}
```

### With options

```go
import (
    "log/slog"
    "os"

    godidit "github.com/TranVuGiang/go_didit"
)

logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

client, err := godidit.NewClient(
    "your-api-key",
    godidit.WithDebug(true),
    godidit.WithLogger(logger),
    godidit.WithBaseURL("https://verification.didit.me"), // optional, this is the default
)
```

### Create a KYC session

```go
resp, err := client.CreateSession(ctx, &godidit.CreateSessionRequest{
    WorkflowID: "your-workflow-id",   // required
    VendorData: "internal-user-id",   // optional
    Callback:   "https://yourapp.com/webhook", // optional
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.SessionID) // redirect the user to resp.URL to start verification
fmt.Println(resp.URL)
```

### Retrieve session decision

```go
decision, err := client.RetrieveSession(ctx, "session-id")
if err != nil {
    log.Fatal(err)
}

fmt.Println(decision.Status)
```

### Error handling

API errors are returned as `*godidit.APIError` and can be inspected with `errors.As`:

```go
decision, err := client.RetrieveSession(ctx, sessionID)
if err != nil {
    var apiErr *godidit.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error: status=%d message=%s\n", apiErr.StatusCode, apiErr.Message)
    } else {
        fmt.Printf("transport error: %v\n", err)
    }
}
```

## API Reference

| Method | Description |
|--------|-------------|
| `CreateSession(ctx, req)` | Create a new KYC verification session |
| `RetrieveSession(ctx, sessionID)` | Retrieve the decision for a session |

## Client Options

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Override the default base URL |
| `WithLogger(logger)` | Set a custom structured logger (`log/slog` compatible) |
| `WithDebug(bool)` | Enable debug logging of requests and responses |
| `WithHTTPClient(client)` | Inject a custom `http.Client` (useful for tests or proxies) |

## Testing

Integration tests require a valid API key. Set the following environment variables before running:

```bash
# Required for all tests
export DIDIT_API_KEY=your-api-key

# Required for TestCreateSession
export DIDIT_WORKFLOW_ID=your-workflow-id

# Required for TestRetrieveSession
export DIDIT_SESSION_ID=your-session-id
```

```bash
go test -v -race ./...
```

Tests automatically skip when the required environment variables are not set.

## License

MIT
