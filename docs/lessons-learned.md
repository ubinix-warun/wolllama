# Building Wolllama on Walrus — Lessons Learned

## What We Built

Wolllama is a decentralized model registry for Ollama. Models are pushed to and pulled from
Walrus decentralized storage. Users can browse, discover, and submit models through a web
registry. Two storage providers: direct Walrus and Tatum's managed gateway.

## The Good

### Walrus Just Works for Raw Blobs

Direct Walrus storage is straightforward. The publisher/aggregator HTTP API is clean —
`PUT /v1/blobs` to store, `GET /v1/blobs/{id}` to retrieve. Content-addressed blobs
return exactly what you uploaded. SHA256 verification passes every time. No surprises.

The `walrus-go` SDK (community-maintained) worked well for basic operations, and we
ended up writing our own thin wrapper around it to add quilt-patch fallback logic.

### Chunked Storage is Necessary

Walrus aggregator caps single-blob downloads at 500 MB. For AI models that regularly
exceed 1 GB, chunking is essential. We split blobs > 256 MB into sequential chunks,
each stored as a separate Walrus object. This also enables partial retries and
parallel downloads in future versions.

### Quilt-Patch IDs Saved Us

Tatum stores blobs using Walrus's quilt (erasure-coded) mechanism. The regular
`/v1/blobs/{id}` endpoint returns a binary wrapper (`\x01` header) for quilt-encoded
blobs. But `/v1/blobs/by-quilt-id/{id}/blob` returns the raw content. This
two-tier endpoint system — while discoverable only through trial and error —
proved essential for cross-provider compatibility.

## The Hard Parts

### Tatum Blob Wrapping — The Biggest Surprise

**Tatum's storage API does not store raw blobs on Walrus.** When you upload a file
through Tatum's `POST /v4/data/storage/upload`, the file is stored on Walrus wrapped
in a Tatum-specific binary encoding. The regular Walrus endpoint returns this wrapper,
not your original bytes.

This took days to debug. The symptom: `wolllama pull` would fetch the manifest,
parse it, download individual blobs, and then SHA256 checksum verification would fail
with completely different hashes. The downloaded bytes were the Tatum wrapper, not
the original content.

**Why this matters:** If you were a user who uploaded through Tatum and shared the
blob ID with someone using direct Walrus, they couldn't read your data. The blob ID
alone is insufficient — you need to know it's a Tatum blob and use the quilt endpoints.

### Discovering the Quilt Endpoints

Tatum's API documentation mentions "Walrus" but doesn't explain the blob ID
discrepancy. We discovered the solution through:

1. **Manual curl exploration** — fetched the certification response and found
   `quiltPatchId`, `quiltId`, `suiObjectId`, and download URLs
2. **Trial and error** — `/v1/blobs/by-quilt-id/{blobId}/blob` returned
   raw content while `/v1/blobs/{blobId}` returned the wrapper
3. **Pattern matching** — quilt-patch IDs had `BAQAEAA` suffixes
   appended to blob IDs

### Three-Tier Fallback Required

To handle all blob types transparently, we needed three download strategies:

```
1. GET /v1/blobs/{id}                    — native Walrus blobs
2. GET /v1/blobs/by-quilt-patch-id/{id}  — Tatum quiltPatchIds (with suffix)
3. GET /v1/blobs/by-quilt-id/{id}/blob   — Tatum blobIds (without suffix)
```

And we detect Tatum wrappers by checking if the first byte is `\x01`. Not elegant,
but it works.

### Multipart Upload Overhead

Tatum's upload API uses `multipart/form-data`. For the 50 MiB upload limit, the
multipart envelope adds overhead, so we had to chunk at 45 MiB to stay under the
limit. This is a hidden constraint not documented in the API reference.

### No Tatum Download Endpoint

Tatum's storage API is write-only — there's no `GET` endpoint to download files.
Downloads must go through the Walrus aggregator directly. This means:
- Tatum API key is only needed for push, not pull
- Pull must know the Walrus network (mainnet/testnet)
- The user experience is asymmetric: push via Tatum, pull via Walrus

## What We'd Do Differently

1. **Test blob retrieval immediately after upload** — don't assume stored bytes
   equal uploaded bytes. Verify with a read-back and SHA256 check.

2. **Request a raw blob upload option from Tatum** — an endpoint that stores
   raw bytes without the binary wrapper would make the gateway truly transparent.

3. **Build chunking from day one** — the 500 MB download limit would have
   been a showstopper for AI models otherwise.

4. **Document the quilt endpoint discovery process** — other builders will
   hit the same wrapper issue. The quilt endpoints should be documented
   alongside the standard blob endpoint.

## Summary

| Aspect | Rating | Notes |
|--------|--------|-------|
| Walrus direct API | ⭐⭐⭐⭐⭐ | Clean, predictable, content-addressed |
| walrus-go SDK | ⭐⭐⭐⭐ | Solid for basics, needed wrapper for quilt |
| Tatum storage gateway | ⭐⭐⭐ | Useful for key management, but blob wrapping adds complexity |
| Tatum API docs | ⭐⭐ | Missing quilt endpoint docs, download endpoint, size constraints |
| Blob size limits | ⭐⭐⭐ | 500 MB download cap means chunking is mandatory for models |
| Overall dev experience | ⭐⭐⭐⭐ | Steep learning curve on Tatum integration, but Walrus itself is solid |
