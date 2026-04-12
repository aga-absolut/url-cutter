package errs

import "errors"

var (
	ErrInGettingUserID      = errors.New("getting id")
	ErrInGettingURLs        = errors.New("getting urls")
	ErrEmptyBody            = errors.New("empty body")
	ErrEmptyBatch           = errors.New("empty batch")
	ErrURLAlreadyExists     = errors.New("url already exists")
	ErrTrustedSubnetIsEmpty = errors.New("empty trustedSubnet")
	ErrInGettingUrlsCount   = errors.New("пetting Urls Count")
	ErrNilIP                = errors.New("nil ip")
	ErrIPNotAllowed         = errors.New("ip not allowed")
)
