package signup

import (
	"context"
	"fmt"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mantil-io/mantil.go"
)

const (
	registrationPartition = "registration"
	activationPartition   = "activation"
	keysPartition         = "keys"
)

var (
	errInternal   = fmt.Errorf("internal server error")
	errBadRequest = fmt.Errorf("bad request")
)

type RegisterRequest struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func (r *RegisterRequest) Valid() bool {
	_, err := mail.ParseAddress(r.Email)
	return r.Email != "" && err == nil
}

type RegisterRecord struct {
	ActivationCode string
	Activations    []string
	RemoteIP       string
	CreatedAt      int64
	Email          string
	Name           string
}

const (
	SourceCli      = 1
	SourceTypeform = 2
)

func (r *RegisterRequest) ToRecord(ip string, raw []byte) RegisterRecord {
	code := uuid.NewString()
	return RegisterRecord{
		ActivationCode: code,
		Email:          r.Email,
		Name:           r.Name,
		CreatedAt:      time.Now().UnixMilli(),
		RemoteIP:       ip,
	}
}

type ActivateRequest struct {
	ActivationCode string `json:"activationCode,omitempty"`
}

func NewActivateRequest(activationCode, workspaceID string) ActivateRequest {
	return ActivateRequest{
		ActivationCode: activationCode,
	}
}

func (r *ActivateRequest) Valid() bool {
	return r.ActivationCode != ""
}

func (r *ActivateRequest) ToRecord(remoteIP string) ActivateRecord {
	return ActivateRecord{
		ID:             uuid.NewString(),
		ActivationCode: r.ActivationCode,
		RemoteIP:       remoteIP,
		CreatedAt:      time.Now().UnixMilli(),
	}
}

type ActivateRecord struct {
	ID             string
	ActivationCode string
	Token          string
	RemoteIP       string
	CreatedAt      int64
}

func (r ActivateRecord) ToTokenClaims() TokenClaims {
	return TokenClaims{
		ActivationCode: r.ActivationCode,
		ActivationID:   r.ID,
		CreatedAt:      time.Now().UnixMilli(),
	}
}

type VerifyRequest struct {
	Token string `json:"token,omitempty"`
}

type Signup struct {
	kv *kv
}

func New() *Signup {
	return &Signup{
		kv: &kv{},
	}
}

func (r *Signup) register(ctx context.Context, req RegisterRequest) (*RegisterRecord, error) {
	if !req.Valid() {
		return nil, errBadRequest
	}
	rec := req.ToRecord(remoteIPRawRequest(ctx))
	if err := r.kv.Registrations().Put(rec.ActivationCode, rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Signup) Register(ctx context.Context, req RegisterRequest) error {
	rec, err := r.register(ctx, req)
	if err != nil {
		return err
	}
	if err := r.sendActivationCode(rec.Email, rec.Name, rec.ActivationCode); err != nil {
		return errInternal
	}
	return nil
}

func (r *Signup) activate(ctx context.Context, req ActivateRequest) (*ActivateRecord, *RegisterRecord, error) {
	if !req.Valid() {
		return nil, nil, errBadRequest
	}

	var rr RegisterRecord
	if err := r.kv.Registrations().Get(req.ActivationCode, &rr); err != nil {
		log.Printf("register record not found for %s, error: %s", req.ActivationCode, err)
		return nil, nil, fmt.Errorf("activation code not found")
	}

	ar := req.ToRecord(remoteIP(ctx))
	token, err := r.generateToken(ar.ToTokenClaims())
	if err != nil {
		log.Printf("failed to encode user token error: %s", err)
		return nil, nil, errInternal
	}
	ar.Token = token

	if err := r.kv.Activations().Put(ar.ID, ar); err != nil {
		return nil, nil, err
	}
	rr.Activations = append(rr.Activations, ar.ID)
	if err := r.kv.Registrations().Put(rr.ActivationCode, rr); err != nil {
		return nil, nil, err
	}
	return &ar, &rr, nil
}

func (r *Signup) Activate(ctx context.Context, req ActivateRequest) (string, error) {
	ar, rr, err := r.activate(ctx, req)
	if err != nil {
		return "", err
	}

	if err := r.sendWelcomeMail(rr.Email, rr.Name); err != nil {
		log.Printf("failed to send welcome mail error %s", err)
	}
	return ar.Token, nil
}

func (r *Signup) Verify(ctx context.Context, req VerifyRequest) (*TokenClaims, error) {
	jwt := strings.TrimSpace(req.Token)
	var ut TokenClaims
	err := r.decodeToken(jwt, &ut)
	if err != nil {
		return nil, fmt.Errorf("invalid token %s", err)
	}
	var ar ActivateRecord
	if err := r.kv.Activations().Get(ut.ActivationID, &ar); err != nil {
		return nil, fmt.Errorf("failed to retrieve activation record for token - %s", err)
	}
	return &ut, nil
}

func remoteIP(ctx context.Context) string {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return ""
	}
	return rc.Request.RemoteIP()
}

func remoteIPRawRequest(ctx context.Context) (string, []byte) {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return "", nil
	}
	return rc.Request.RemoteIP(), rc.Request.Raw
}
