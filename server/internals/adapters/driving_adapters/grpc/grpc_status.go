package adapters

import (
	domainErrors "github.com/cybrarymin/gRPC/server/internals/domains/errors"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StatusCheck(err interface{}) error {
	switch e := err.(type) {
	case map[string]string:
		violation := []*errdetails.BadRequest_FieldViolation{}
		for k, value := range e {
			violation = append(violation, &errdetails.BadRequest_FieldViolation{
				Field:       k,
				Description: value,
			})
		}
		nerrDetail := &errdetails.BadRequest{
			FieldViolations: violation,
		}
		st := status.New(codes.InvalidArgument, "invalid input")
		stwithdetails, attacherr := st.WithDetails(nerrDetail)
		if attacherr != nil {
			return status.Error(codes.Internal, "couldn't attach error details to the status")
		}
		return stwithdetails.Err()

	case error:
		switch {
		case domainErrors.IsAlreadyExists(e):
			return status.Error(codes.AlreadyExists, e.Error())
		case domainErrors.IsNotFound(e):
			return status.Error(codes.NotFound, e.Error())
		case domainErrors.IsInvalidInput(e):
			return status.Error(codes.InvalidArgument, e.Error())
		default:
			return status.Error(codes.Internal, e.Error())
		}
	}
	return nil
}
