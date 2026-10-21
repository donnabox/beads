### Advertised limits

The Read discovery document's `limits` uses the closed `advertisedLimits`
definition. Its Read-applicable groups are `page`, `request`, `resource`,
`selector`, and `retention.maximumSnapshotLifetime`; the schema admits no
`validation` group. Schema acceptance of a limit does not expose its capability.

The discovery document MAY contain a `limits` object. The object is optional
so that a small implementation can expose a conforming profile without
predicting every operational bound. Omission means only that the bound is not
pre-advertised; it does not mean infinite capacity and does not permit silent
truncation or a non-normative failure response.

When present, `limits` is divided into capability groups. A group is relevant
only when the advertised profile exposes that capability. Each advertised
value is a binding positive integer or ISO 8601 duration:

- `page.defaultItems` and `page.maximumItems` count Resource records;
- `request.targetBytes` counts octets in the encoded HTTP request target, and
  `request.bodyBytes` counts octets in the representation body;
- `resource.representationBytes` and `resource.propertiesBytes` count UTF-8
  bytes in the corresponding JSON serialization; metadata is covered by the
  complete representation bound, not by the properties-only bound;
- `selector.bytes` counts UTF-8 bytes of the decoded Selector string: after
  percent-decoding for query parameters, or JSON string decoding for a
  Selector in a request body, without additional URL decoding;
  `selector.depth` and `selector.nodes` count parsed Selector structure;
- `retention.maximumSnapshotLifetime` bounds the lifetime of a pagination
  continuation when advertised.

Fields and groups not advertised carry no implicit numeric value. A client
may use advertised values for request planning. Conformance tests may probe
them and require the server to enforce the advertised boundary consistently.
For example:

```json
{
  "limits": {
    "page": {
      "defaultItems": 50,
      "maximumItems": 200
    },
    "request": {
      "targetBytes": 2048,
      "bodyBytes": 65536
    }
  }
}
```

