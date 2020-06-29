package restutil

// StatusType determines if the operation status type
// is "Success" or "Failure".
type StatusType string

const (
	// StatusSuccess is "Success" status
	StatusSuccess StatusType = "Success"
	// StatusFailure is "Failure" status
	StatusFailure StatusType = "Failure"
)

// StatusReason is a machine-readable description of why this
// operation is in the "Failure" status.
type StatusReason string

const (
	// StatusReasonUnknown means the server has declined to indicate a specific reason.
	// The details field may contain other information about this error.
	// Status code 500.
	StatusReasonUnknown StatusReason = ""

	// StatusReasonUnauthorized means the server can be reached and understood the request, but requires
	// the user to present appropriate authorization credentials (identified by the WWW-Authenticate header)
	// in order for the action to be completed. If the user has specified credentials on the request, the
	// server considers them insufficient.
	// Status code 401
	StatusReasonUnauthorized StatusReason = "Unauthorized"

	// StatusReasonForbidden means the server can be reached and understood the request, but refuses
	// to take any further action.  It is the result of the server being configured to deny access for some reason
	// to the requested resource by the client.
	// Status code 403
	StatusReasonForbidden StatusReason = "Forbidden"

	// StatusReasonNotFound means one or more resources required for this operation
	// could not be found.
	// Status code 404
	StatusReasonNotFound StatusReason = "NotFound"

	// StatusReasonAlreadyExists means the resource you are creating already exists.
	// Status code 409
	StatusReasonAlreadyExists StatusReason = "AlreadyExists"

	// StatusReasonConflict means the requested operation cannot be completed
	// due to a conflict in the operation. The client may need to alter the
	// request. Each resource may define custom details that indicate the
	// nature of the conflict.
	// Status code 409
	StatusReasonConflict StatusReason = "Conflict"

	// StatusReasonGone means the item is no longer available at the server and no
	// forwarding address is known.
	// Status code 410
	StatusReasonGone StatusReason = "Gone"

	// StatusReasonInvalid means the requested create or update operation cannot be
	// completed due to invalid data provided as part of the request. The client may
	// need to alter the request. When set, the client may use the StatusDetails
	// message field as a summary of the issues encountered.
	// Status code 422
	StatusReasonInvalid StatusReason = "Invalid"

	// StatusReasonServerTimeout means the server can be reached and understood the request,
	// but cannot complete the action in a reasonable time. The client should retry the request.
	// This is may be due to temporary server load or a transient communication issue with
	// another server. Status code 500 is used because the HTTP spec provides no suitable
	// server-requested client retry and the 5xx class represents actionable errors.
	// Status code 500
	StatusReasonServerTimeout StatusReason = "ServerTimeout"

	// StatusReasonTimeout means that the request could not be completed within the given time.
	// The server cannot complete the operation within a reasonable amount of time.
	// Status code 504
	StatusReasonTimeout StatusReason = "Timeout"

	// StatusReasonTooManyRequests means the server experienced too many requests within a
	// given window and that the client must wait to perform the action again. A client may
	// always retry the request that led to this error.
	// Status code 429
	StatusReasonTooManyRequests StatusReason = "TooManyRequests"

	// StatusReasonBadRequest means that the request itself was invalid, because the request
	// doesn't make any sense, for example deleting a read-only object.  This is different than
	// StatusReasonInvalid above which indicates that the API call could possibly succeed, but the
	// data was invalid.  API calls that return BadRequest can never succeed.
	// Status code 400
	StatusReasonBadRequest StatusReason = "BadRequest"

	// StatusReasonMethodNotAllowed means that the action the client attempted to perform on the
	// resource was not supported by the code - for instance, attempting to delete a resource that
	// can only be created. API calls that return MethodNotAllowed can never succeed.
	// Status code 405
	StatusReasonMethodNotAllowed StatusReason = "MethodNotAllowed"

	// StatusReasonNotAcceptable means that the accept types indicated by the client were not acceptable
	// to the server - for instance, attempting to receive protobuf for a resource that supports only json and yaml.
	// API calls that return NotAcceptable can never succeed.
	// Status code 406
	StatusReasonNotAcceptable StatusReason = "NotAcceptable"

	// StatusReasonRequestEntityTooLarge means that the request entity is too large.
	// Status code 413
	StatusReasonRequestEntityTooLarge StatusReason = "RequestEntityTooLarge"

	// StatusReasonUnsupportedMediaType means that the content type sent by the client is not acceptable
	// to the server - for instance, attempting to send protobuf for a resource that supports only json and yaml.
	// API calls that return UnsupportedMediaType can never succeed.
	// Status code 415
	StatusReasonUnsupportedMediaType StatusReason = "UnsupportedMediaType"

	// StatusReasonInternalError indicates that an internal error occurred, it is unexpected
	// and the outcome of the call is unknown.
	// Status code 500
	StatusReasonInternalError StatusReason = "InternalError"

	// StatusReasonExpired indicates that the request is invalid because the content you are requesting
	// has expired and is no longer available. It is typically associated with watches that can't be
	// serviced.
	// Status code 410 (gone)
	StatusReasonExpired StatusReason = "Expired"

	// StatusReasonServiceUnavailable means that the request itself was valid,
	// but the requested service is unavailable at this time.
	// Retrying the request after some time might succeed.
	// Status code 503
	StatusReasonServiceUnavailable StatusReason = "ServiceUnavailable"
)

// Status is a return value for calls that don't return
// other objects.
type Status struct {
	// Status of the operation.
	// One of: "Success" or "Failure".
	Status StatusType `json:"status,omitempty"`
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty"`
	// A machine-readable description of why this operation is in the
	// "Failure" status. If this value is empty there
	// is no information available. A Reason clarifies an HTTP status
	// code but does not override it.
	Reason StatusReason `json:"reason,omitempty"`
	// Extended data associated with the reason. Each reason may define its
	// own extended details. This field is optional and the data returned
	// is not guaranteed to conform to any schema except that defined by
	// the reason type.
	Details interface{} `json:"details,omitempty"`
	// Suggested HTTP return code for this status, 0 if not set.
	Code int `json:"code,omitempty"`
}

// StatusError is a Status of type error
type StatusError Status

func (status *StatusError) Error() string {
	return status.Message
}

// Error creates new failure status with message and StatusReason.
func Error(message string, reason StatusReason) *StatusError {

	var code int

	switch reason {
	case StatusReasonUnknown:
		code = 500
	case StatusReasonUnauthorized:
		code = 401
	case StatusReasonForbidden:
		code = 403
	case StatusReasonNotFound:
		code = 404
	case StatusReasonAlreadyExists:
		code = 409
	case StatusReasonConflict:
		code = 409
	case StatusReasonGone:
		code = 410
	case StatusReasonInvalid:
		code = 422
	case StatusReasonServerTimeout:
		code = 500
	case StatusReasonTimeout:
		code = 504
	case StatusReasonTooManyRequests:
		code = 429
	case StatusReasonBadRequest:
		code = 400
	case StatusReasonMethodNotAllowed:
		code = 405
	case StatusReasonNotAcceptable:
		code = 406
	case StatusReasonRequestEntityTooLarge:
		code = 413
	case StatusReasonUnsupportedMediaType:
		code = 415
	case StatusReasonInternalError:
		code = 500
	case StatusReasonExpired:
		code = 410
	case StatusReasonServiceUnavailable:
		code = 503
	}

	status := StatusError{
		Status:  StatusFailure,
		Message: message,
		Reason:  reason,
		Code:    code,
	}

	return &status
}

// ErrorWithDetails creates new failure status with message, StatusReason and details
func ErrorWithDetails(message string, reason StatusReason, details interface{}) *StatusError {
	errStatus := Error(message, reason)
	errStatus.Details = details
	return errStatus
}