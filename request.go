package grab

import (
	"context"
	"hash"
	"net/http"
	"net/url"
)

// A Request represents an HTTP file transfer request to be sent by a Client.
type Request struct {
	// Label is an arbitrary string which may used to label a Request with a
	// user friendly name.
	Label string

	// Tag is an arbitrary interface which may be used to relate a Request to
	// other data.
	Tag interface{}

	// HTTPRequest specifies the http.Request to be sent to the remote server to
	// initiate a file transfer. It includes request configuration such as URL,
	// protocol version, HTTP method, request headers and authentication.
	HTTPRequest *http.Request

	// Filename specifies the path where the file transfer will be stored in
	// local storage. If Filename is empty or a directory, the true Filename will
	// be resolved using Content-Disposition headers or the request URL.
	//
	// An empty string means the transfer will be stored in the current working
	// directory.
	Filename string

	// SkipExisting specifies that ErrFileExists should be returned if the
	// destination path already exists. The existing file will not be checked for
	// completeness.
	SkipExisting bool

	// NoResume specifies that a partially completed download will be restarted
	// without attempting to resume any existing file. If the download is already
	// completed in full, it will not be restarted.
	NoResume bool

	// NoCreateDirectories specifies that any missing directories in the given
	// Filename path should not be created automatically, if they do not already
	// exist.
	NoCreateDirectories bool

	// RemoteTime specifies that grab should try to determine the timestamp of the
	// remote file and apply it to the local copy.
	RemoteTime bool

	// Size specifies the expected size of the file transfer if known. If the
	// server response size does not match, the transfer is cancelled and
	// ErrBadLength returned.
	Size int64

	// BufferSize specifies the size in bytes of the buffer that is used for
	// transferring the requested file. Larger buffers may result in faster
	// throughput but will use more memory and result in less frequent updates
	// to the transfer progress statistics. Default: 32KB.
	BufferSize int

	// hash, checksum and deleteOnError - set via SetChecksum.
	hash          hash.Hash
	checksum      []byte
	deleteOnError bool

	// Context for cancellation and timeout - set via WithContext
	ctx context.Context
}

// NewRequest returns a new file transfer Request suitable for use with
// Client.Do.
func NewRequest(dst, urlStr string) (*Request, error) {
	if dst == "" {
		dst = "."
	}

	// create http request
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	return &Request{
		HTTPRequest: req,
		Filename:    dst,
	}, nil
}

// Context returns the request's context. To change the context, use
// WithContext.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// The context controls cancelation.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}

	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx

	// propagate to HTTPRequest
	r2.HTTPRequest = r2.HTTPRequest.WithContext(ctx)
	return r2
}

// URL returns the URL to be downloaded.
func (r *Request) URL() *url.URL {
	return r.HTTPRequest.URL
}

// SetChecksum sets the desired hashing algorithm and checksum value to validate
// a downloaded file. Once the download is complete, the given hashing algorithm
// will be used to compute the actual checksum of the downloaded file. If the
// checksums do not match, an error will be returned by the associated
// Response.Err method.
//
// If deleteOnError is true, the downloaded file will be deleted automatically
// if it fails checksum validation.
//
// To prevent corruption of the computed checksum, the given hash must not be
// used by any other request or goroutines.
//
// To disable checksum validation, call SetChecksum with a nil hash.
func (r *Request) SetChecksum(h hash.Hash, sum []byte, deleteOnError bool) {
	if h == nil {
		h, sum, deleteOnError = nil, nil, false
	}

	r.hash = h
	r.checksum = sum
	r.deleteOnError = deleteOnError
}
