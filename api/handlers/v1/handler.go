package v1

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	jwtg "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	validate "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbo "gitlab.udevs.io/macbro/mb_admin_api_gateway/genproto/order_service"

	"gitlab.udevs.io/macbro/mb_admin_api_gateway/api/models"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/config"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/mb_variables"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/pkg/grpc_client"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/pkg/jwt"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/pkg/logger"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/services"
	"gitlab.udevs.io/macbro/mb_admin_api_gateway/storage/repo"
)

type handlerV1 struct {
	log        logger.Logger
	grpcClient *grpc_client.GrpcClient
	cfg        config.Config
	services   services.ServiceManager
	redis      redis.Conn
}

//HandlerV1Config ...
type HandlerV1Config struct {
	Logger     logger.Logger
	GrpcClient *grpc_client.GrpcClient
	Cfg        config.Config
	Services   services.ServiceManager
	Redis      redis.Conn
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
		services:   c.Services,
		redis:      c.Redis,
	}
}

// CompareOrderItems ...
func CompareOrderItems(old, new []*pbo.OrderItem) bool {
	var flag = false
	if len(old) != len(new) {
		return false
	}
	for _, newItem := range new {
		for _, oldItem := range old {
			if newItem.ProductId == oldItem.ProductId {
				flag = true
			}
			if newItem.Quantity == oldItem.Quantity {
				flag = true
			}
		}
	}
	return flag
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

func (h *handlerV1) HandleBadRequest(c *gin.Context, err error, message string) {
	h.log.Error(message, logger.Error(err))
	c.JSON(http.StatusBadRequest, mb_variables.Error{
		Code:    http.StatusBadRequest,
		Message: message,
		Reason:  err.Error(),
	})
}
func (h *handlerV1) HandleInternalServerError(c *gin.Context, err error, message string) {
	c.JSON(http.StatusInternalServerError, mb_variables.Error{
		Code:    http.StatusInternalServerError,
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

func handleError(log logger.Logger, c *gin.Context, err error, message string) (hasError bool) {
	if err != nil {
		log.Error("Error: ", logger.Error(err))
	}
	st, ok := status.FromError(err)

	if st.Code() == codes.AlreadyExists || st.Code() == codes.InvalidArgument {
		log.Error(message+", already exists", logger.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   ErrorCodeAlreadyExists,
		})
		return
	} else if st.Code() == codes.NotFound {
		log.Error(message+", not found", logger.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   ErrorCodeNotFound,
		})
		return
	} else if st.Code() == codes.Unavailable {
		log.Error(message+", service unavailable", logger.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   ErrorCodeNotFound,
		})
		return
	} else if !ok || st.Code() == codes.Internal || st.Code() == codes.Unknown || err != nil {
		log.Error(message+", internal server error", logger.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   ErrorCodeInternal,
		})
		return
	}
	return true
}

func (h *handlerV1) MakeProxy(c *gin.Context, proxyUrl, path string) (err error) {
	req := c.Request

	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		h.log.Error("error in parse addr: %v", logger.Error(err))
		c.String(http.StatusInternalServerError, "error")
		return
	}
	req.URL.Scheme = proxy.Scheme
	req.URL.Host = proxy.Host
	req.URL.Path = path
	transport := http.DefaultTransport
	// req.URL.RawQuery = "name=string"
	resp, err := transport.RoundTrip(req)
	if !handleError(h.log, c, err, "error in round trip:") {
		return
	}

	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Header(k, v)
		}
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)
	_, _ = bufio.NewReader(resp.Body).WriteTo(c.Writer)
	return
}

func (h *handlerV1) MakeProxyValidator(c *gin.Context, proxyUrl, path, searchKey, searchValue string) (err error) {
	req := c.Request

	proxy, err := url.Parse(proxyUrl)
	if err != nil {
		h.log.Error("error in parse addr: %v", logger.Error(err))
		c.String(http.StatusInternalServerError, "error")
		return
	}
	req.URL.Path = path
	req.Method = "GET"
	req.URL.RawQuery = fmt.Sprintf("%v=%v", searchKey, searchValue)
	// fmt.Printf(fmt.Sprintf("%v=%v", searchKey, searchValue))
	req.URL.Scheme = proxy.Scheme
	req.URL.Host = proxy.Host
	transport := http.DefaultTransport
	fmt.Println(req)
	resp, err := transport.RoundTrip(req)
	fmt.Println(resp.Body)
	if !handleError(h.log, c, err, "error in round trip:") {
		return
	}
	defer resp.Body.Close()

	c.Status(resp.StatusCode)

	return
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

func handleGRPCErr(c *gin.Context, l logger.Logger, err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	var errI interface{} = models.InternalServerError{
		Code:    ErrorCodeInternal,
		Message: "Internal Server Error",
	}
	httpCode := http.StatusInternalServerError
	if ok && st.Code() == codes.InvalidArgument {
		httpCode = http.StatusBadRequest
		errI = ErrorBadRequest
	}
	c.JSON(httpCode, models.ResponseError{
		Error: errI,
	})
	if ok {
		l.Error(fmt.Sprintf("code=%d message=%s", st.Code(), st.Message()), logger.Error(err))
	}
	return true
}

func writeMessageAsJSON(c *gin.Context, l logger.Logger, msg proto.Message) {
	if msg == nil {
		c.String(http.StatusOK, "")
		return
	}
	var jspbMarshal jsonpb.Marshaler

	jspbMarshal.OrigName = true
	jspbMarshal.EmitDefaults = true

	js, err := jspbMarshal.MarshalToString(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeInternal,
				Message: "Internal Server Error",
			},
		})
		l.Error("Error while marshaling", logger.Error(err))
		return
	}
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, js)
}

func handleGrpcErrWithMessage(c *gin.Context, l logger.Logger, err error, message string) bool {
	st, ok := status.FromError(err)
	if !ok || st.Code() == codes.Internal {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeInternal,
				Message: st.Message(),
			},
		})
		l.Error(message, logger.Error(err))
		return true
	}
	if st.Code() == codes.NotFound {
		c.JSON(http.StatusNotFound, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeNotFound,
				Message: st.Message(),
			},
		})
		l.Error(message+", not found", logger.Error(err))
		return true
	} else if st.Code() == codes.Unavailable {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeInternal,
				Message: "Internal Server Error",
			},
		})
		l.Error(message+", service unavailable", logger.Error(err))
		return true
	} else if st.Code() == codes.AlreadyExists {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeAlreadyExists,
				Message: st.Message(),
			},
		})
		l.Error(message+", already exists", logger.Error(err))
		return true
	} else if st.Code() == codes.InvalidArgument {
		c.JSON(http.StatusBadRequest, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorBadRequest,
				Message: st.Message(),
			},
		})
		l.Error(message+", invalid field", logger.Error(err))
		return true
	} else if st.Code() == codes.Code(20) {
		c.JSON(http.StatusBadRequest, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorBadRequest,
				Message: st.Message(),
			},
		})
		l.Error(message+", invalid field", logger.Error(err))
		return true
	} else if st.Err() != nil {
		c.JSON(http.StatusBadRequest, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorBadRequest,
				Message: st.Message(),
			},
		})
		l.Error(message+", invalid field", logger.Error(err))
		return true
	}
	return false
}

func handleInternalWithMessage(c *gin.Context, l logger.Logger, err error, message string) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeInternal,
				Message: "Internal Server Error",
			},
		})
		l.Error(message, logger.Error(err))
		return true
	}

	return false
}

func handleStorageErrWithMessage(c *gin.Context, l logger.Logger, err error, message string) bool {
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeNotFound,
				Message: "Not found",
			},
		})
		l.Error(message+", not found", logger.Error(err))
		return true
	} else if err == repo.ErrAlreadyExists {
		c.JSON(http.StatusBadRequest, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeAlreadyExists,
				Message: "Already Exists",
			},
		})
		l.Error(message+", already exists", logger.Error(err))
		return true
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, models.ResponseError{
			Error: models.InternalServerError{
				Code:    ErrorCodeInternal,
				Message: "Internal Server Error",
			},
		})
		l.Error(message, logger.Error(err))
		return true
	}

	return false
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
			Error: ErrorCodeUnauthorized,
		})
		h.log.Error("Unauthorized request: ", logger.Error(ErrUnauthorized))
		return nil, ErrUnauthorized
	}

	claims, err = jwt.ExtractClaims(authorization.Token, signingKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ResponseError{
			Error: ErrorCodeUnauthorized,
		})
		h.log.Error("Unauthorized request: ", logger.Error(err))
		return nil, ErrUnauthorized
	}

	return claims, nil
}
