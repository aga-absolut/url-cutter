package errs

import "errors"

var (
	ErrInGettingUserID      = errors.New("error in getting id")
	ErrInGettingURLs        = errors.New("error in getting urls")
	ErrEmptyBody            = errors.New("error empty body")
	ErrEmptyBatch           = errors.New("error empty batch")
	ErrURLAlreadyExists     = errors.New("url already exists")
	ErrTrustedSubnetIsEmpty = errors.New("error empty trustedSubnet")
	ErrInGettingUrlsCount   = errors.New("error In Getting Urls Count")
	ErrNilIP                = errors.New("error nil ip")
	ErrIPNotAllowed         = errors.New("error ip not allowed")
)
