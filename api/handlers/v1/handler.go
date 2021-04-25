package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"github.com/abdukhashimov/go_gin_example/api/models"
	"github.com/abdukhashimov/go_gin_example/config"
	"github.com/abdukhashimov/go_gin_example/pkg/grpc_client"
	"github.com/abdukhashimov/go_gin_example/pkg/jwt"
	"github.com/abdukhashimov/go_gin_example/pkg/logger"
	jwtg "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	validate "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/bcrypt"
)

type handlerV1 struct {
	log        logger.Logger
	grpcClient *grpc_client.GrpcClient
	cfg        config.Config
}

//HandlerV1Config ...
type HandlerV1Config struct {
	Logger     logger.Logger
	GrpcClient *grpc_client.GrpcClient
	Cfg        config.Config
}

const (
	//ErrorCodeInvalidURL ...
	ErrorCodeInvalidURL = "INVALID_URL"
	//ErrorCodeInvalidJSON ...
	ErrorCodeInvalidJSON = "INVALID_JSON"
	//ErrorCodeInternal ...
	ErrorCodeInternal = "INTERNAL"
	//ErrorCodeUnauthorized ...
	ErrorCodeUnauthorized = "UNAUTHORIZED"
	//ErrorCodeAlreadyExists ...
	ErrorCodeAlreadyExists = "ALREADY_EXISTS"
	//ErrorCodeNotFound ...
	ErrorCodeNotFound = "NOT_FOUND"
	//ErrorCodeInvalidCode ...
	ErrorCodeInvalidCode = "INVALID_CODE"
	//ErrorBadRequest ...
	ErrorBadRequest = "BAD_REQUEST"
	//ErrorCodeForbidden ...
	ErrorCodeForbidden = "FORBIDDEN"
	//ErrorCodeNotApproved ...
	ErrorCodeNotApproved = "NOT_APPROVED"
	//ErrorCodeWrongClub ...
	ErrorCodeWrongClub = "WRONG_CLUB"
	//ErrorCodePasswordsNotEqual ...
	ErrorCodePasswordsNotEqual = "PASSWORDS_NOT_EQUAL"
)

var (
	signingKey = []byte("FfLbN7pIEYe8@!EqrttOLiwa(H8)7Ddo")
)

//New ...
func New(c *HandlerV1Config) *handlerV1 {
	return &handlerV1{
		log:        c.Logger,
		grpcClient: c.GrpcClient,
		cfg:        c.Cfg,
	}
}

// ProtoToStruct is ..
func ProtoToStruct(s interface{}, p proto.Message) error {
	var jm jsonpb.Marshaler

	jm.EmitDefaults = true
	jm.OrigName = true

	ms, err := jm.MarshalToString(p)

	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(ms), &s)

	return err
}

func GeneratePasswordHash(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), 10)
}

func (h *handlerV1) handleBadRequest(c *gin.Context, err error, message string) {
	h.log.Error(message, logger.Error(err))
	c.JSON(http.StatusBadRequest, models.ResponseError{
		Message: message,
		Reason:  err.Error(),
	})
}

func (h *handlerV1) handleInternalServerError(c *gin.Context, err error, message string) {
	h.log.Error(message, logger.Error(err))
	c.JSON(http.StatusBadRequest, models.ResponseError{
		Message: message,
		Reason:  err.Error(),
	})
}

func ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return errors.New("phone_number is blank")
	}
	pattern := regexp.MustCompile(`^(\+[0-9]{12})$`)

	if !(pattern.MatchString(phoneNumber) && phoneNumber[0:4] == "+998") {
		return errors.New("phone_number must be +998XXXXXXXXX")
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be blank")
	}
	if len(password) < 5 || len(password) > 30 {
		return errors.New("password length should be 8 to 30 characters")
	}
	if validate.Validate(password, validate.Match(regexp.MustCompile("[0-9]"))) != nil {
		return errors.New("password should contain at least one number")
	}
	if validate.Validate(password, validate.Match(regexp.MustCompile("[A-Za-z]"))) != nil {
		return errors.New("password should contain at least one alphabetic character")
	}
	return nil
}

//ParsePageQueryParam ...
func ParsePageQueryParam(c *gin.Context) (uint64, error) {
	page, err := strconv.ParseUint(c.DefaultQuery("page", "1"), 10, 10)
	if err != nil {
		return 0, err
	}
	if page < 0 {
		return 0, errors.New("page must be an positive integer")
	}
	if page == 0 {
		return 1, nil
	}
	return page, nil
}

//ParsePageSizeQueryParam ...
func ParsePageSizeQueryParam(c *gin.Context) (uint64, error) {
	pageSize, err := strconv.ParseUint(c.DefaultQuery("page_size", "10"), 10, 10)
	if err != nil {
		return 0, err
	}
	if pageSize < 0 {
		return 0, errors.New("page_size must be an positive integer")
	}
	return pageSize, nil
}

//ParseLimitQueryParam ...
func ParseLimitQueryParam(c *gin.Context) (uint64, error) {
	limit, err := strconv.ParseUint(c.DefaultQuery("limit", "10"), 10, 10)
	if err != nil {
		return 0, err
	}
	if limit < 0 {
		return 0, errors.New("page_size must be an positive integer")
	}
	if limit == 0 {
		return 10, nil
	}
	return limit, nil
}

//ParseSearchQueryParam ...
func ParseSearchQueryParam(c *gin.Context) (string, error) {
	s := c.DefaultQuery("search", "")
	return s, nil
}

//ParseActiveQueryParam ...
func ParseActiveQueryParam(c *gin.Context) (bool, error) {
	a, err := strconv.ParseBool(c.DefaultQuery("active", "false"))
	if err != nil {
		return false, err
	}
	return a, nil
}

//ParseInactiveQueryParam ...
func ParseInactiveQueryParam(c *gin.Context) (bool, error) {
	a, err := strconv.ParseBool(c.DefaultQuery("inactive", "false"))
	if err != nil {
		return false, err
	}
	return a, nil
}

//ParseRecommendedQueryParam ...
func ParseRecommendedQueryParam(c *gin.Context) (bool, error) {
	a, err := strconv.ParseBool(c.DefaultQuery("recommended", "false"))
	if err != nil {
		return false, err
	}
	return a, nil
}

//ParseOnlyRelatedQueryParam ...
func ParseOnlyRelatedQueryParam(c *gin.Context) (bool, error) {
	a, err := strconv.ParseBool(c.DefaultQuery("onlyRelatedProducts", "false"))
	if err != nil {
		return false, err
	}
	return a, nil
}

//ParsePopularQueryParam ...
func ParsePopularQueryParam(c *gin.Context) (bool, error) {
	a, err := strconv.ParseBool(c.DefaultQuery("popular", "false"))
	if err != nil {
		return false, err
	}
	return a, nil
}

//FloatToString ...
func FloatToString(inputNum float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(inputNum, 'f', -1, 64)
}

//ParsePositionQueryParam ...
func ParsePositionQueryParam(c *gin.Context) (string, error) {
	s := c.DefaultQuery("position", "")
	return s, nil
}

func userInfo(h *handlerV1, c *gin.Context) (models.UserInfo, error) {
	claims, err := GetClaims(h, c)

	if err != nil {
		return models.UserInfo{}, err
	}

	userID := claims["sub"].(string)
	userRole := claims["role"].(string)

	return models.UserInfo{
		ID:   userID,
		Role: userRole,
	}, nil
}

// GetClaims function for parsing authorization header
func GetClaims(h *handlerV1, c *gin.Context) (jwtg.MapClaims, error) {
	var (
		ErrUnauthorized = errors.New("unauthorized")
		authorization   models.AuthorizationModel
		claims          jwtg.MapClaims
		err             error
	)

	authorization.Token = c.GetHeader("Authorization")
	if c.Request.Header.Get("Authorization") == "" {
		c.JSON(http.StatusUnauthorized, models.ResponseError{
			Message: "You are not authorized to make this request",
			Reason:  ErrorCodeUnauthorized,
		})

		h.log.Error("Unauthorized request: ", logger.Error(ErrUnauthorized))
		return nil, ErrUnauthorized
	}

	claims, err = jwt.ExtractClaims(authorization.Token, signingKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ResponseError{
			Message: "You are not authorized to make this request",
			Reason:  ErrorCodeUnauthorized,
		})

		h.log.Error("Unauthorized request: ", logger.Error(err))
		return nil, ErrUnauthorized
	}

	return claims, nil
}
