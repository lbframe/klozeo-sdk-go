# Klozeo Go SDK

Official Go client for the [Klozeo](https://klozeo.com) API.

## Installation

```bash
go get github.com/lbframe/sdk-go
```

Requires Go 1.23+.

## Quick Start

```go
import "github.com/lbframe/sdk-go"

client := klozeo.New("sk_live_your_api_key")
ctx := context.Background()

// Create a lead
resp, err := client.Create(ctx, &klozeo.Lead{
    Name:   "Acme Corporation",
    Source: "website",
    City:   "San Francisco",
    Email:  "contact@acme.com",
    Rating: klozeo.Ptr(4.5),
    Tags:   []string{"enterprise", "saas"},
})

// List with filters + auto-pagination
for lead, err := range client.Iterator(ctx,
    klozeo.City().Eq("Berlin"),
    klozeo.Sort(klozeo.FieldRating, klozeo.Desc),
) {
    if err != nil { panic(err) }
    fmt.Println(lead.Name)
}

// Export as CSV
reader, _ := client.Export(ctx, klozeo.ExportCSV)
defer reader.Close()
io.Copy(os.Stdout, reader)
```

## Documentation

- Full API reference: <https://docs.klozeo.com>
- GoDoc: <https://pkg.go.dev/github.com/lbframe/sdk-go>
- GitHub: <https://github.com/lbframe/sdk-go>
